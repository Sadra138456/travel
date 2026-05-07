package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

type ClientEngine struct {
	transport *ClientTransport
	enc       *ClientQuantumEncryption
	localAddr string
}

func NewClientEngine(transport *ClientTransport, enc *ClientQuantumEncryption, localAddr string) *ClientEngine {
	return &ClientEngine{
		transport: transport,
		enc:       enc,
		localAddr: localAddr,
	}
}

func (e *ClientEngine) Start(ctx context.Context) error {
	// Connect to server
	if err := e.transport.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer e.transport.Close()

	// Start local SOCKS5 server
	listener, err := net.Listen("tcp", e.localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	fmt.Printf("SOCKS5 proxy listening on %s\n", e.localAddr)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go e.handleConnection(ctx, conn)
		}
	}
}

func (e *ClientEngine) handleConnection(ctx context.Context, localConn net.Conn) {
	defer localConn.Close()

	// SOCKS5 handshake
	if err := e.socks5Handshake(localConn); err != nil {
		return
	}

	// Get target address
	targetAddr, err := e.socks5GetTarget(localConn)
	if err != nil {
		return
	}

	// Open stream to server
	serverConn, err := e.transport.OpenStream(ctx)
	if err != nil {
		return
	}
	defer serverConn.Close()

	// Send target address to server
	if err := e.sendTargetAddr(serverConn, targetAddr); err != nil {
		return
	}

	// Relay data
	e.relay(localConn, serverConn)
}

func (e *ClientEngine) socks5Handshake(conn net.Conn) error {
	buf := make([]byte, 256)

	// Read version and methods
	n, err := conn.Read(buf)
	if err != nil || n < 2 {
		return fmt.Errorf("handshake failed")
	}

	// Send no auth required
	_, err = conn.Write([]byte{0x05, 0x00})
	return err
}

func (e *ClientEngine) socks5GetTarget(conn net.Conn) (string, error) {
	buf := make([]byte, 256)

	n, err := conn.Read(buf)
	if err != nil || n < 7 {
		return "", fmt.Errorf("failed to read request")
	}

	// Parse address
	addrType := buf[3]
	var host string
	var port uint16

	switch addrType {
	case 0x01: // IPv4
		host = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		port = binary.BigEndian.Uint16(buf[8:10])
	case 0x03: // Domain
		addrLen := int(buf[4])
		host = string(buf[5 : 5+addrLen])
		port = binary.BigEndian.Uint16(buf[5+addrLen : 7+addrLen])
	default:
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return "", fmt.Errorf("unsupported address type")
	}

	// Send success
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	return fmt.Sprintf("%s:%d", host, port), nil
}

func (e *ClientEngine) sendTargetAddr(conn net.Conn, addr string) error {
	addrBytes := []byte(addr)
	encrypted, err := e.enc.Encrypt(addrBytes)
	if err != nil {
		return err
	}

	// Send length + encrypted address
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)))

	if _, err := conn.Write(lenBuf); err != nil {
		return err
	}
	if _, err := conn.Write(encrypted); err != nil {
		return err
	}

	return nil
}

func (e *ClientEngine) relay(local, remote net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Local -> Remote (encrypt)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := local.Read(buf)
			if err != nil {
				return
			}

			encrypted, err := e.enc.Encrypt(buf[:n])
			if err != nil {
				return
			}

			lenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)))

			if _, err := remote.Write(lenBuf); err != nil {
				return
			}
			if _, err := remote.Write(encrypted); err != nil {
				return
			}
		}
	}()

	// Remote -> Local (decrypt)
	go func() {
		defer wg.Done()
		for {
			lenBuf := make([]byte, 4)
			if _, err := io.ReadFull(remote, lenBuf); err != nil {
				return
			}

			length := binary.BigEndian.Uint32(lenBuf)
			encrypted := make([]byte, length)
			if _, err := io.ReadFull(remote, encrypted); err != nil {
				return
			}

			decrypted, err := e.enc.Decrypt(encrypted)
			if err != nil {
				return
			}

			if _, err := local.Write(decrypted); err != nil {
				return
			}
		}
	}()

	wg.Wait()
}

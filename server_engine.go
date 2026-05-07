package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

type ServerEngine struct {
	transport *ServerTransport
	enc       *ServerQuantumEncryption
}

func NewServerEngine(transport *ServerTransport, enc *ServerQuantumEncryption) *ServerEngine {
	return &ServerEngine{
		transport: transport,
		enc:       enc,
	}
}

func (e *ServerEngine) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := e.transport.Accept(ctx)
			if err != nil {
				continue
			}
			go e.handleConnection(conn)
		}
	}
}

func (e *ServerEngine) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Read target address
	targetAddr, err := e.readTargetAddr(clientConn)
	if err != nil {
		return
	}

	// Connect to target
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer targetConn.Close()

	// Relay data
	e.relay(clientConn, targetConn)
}

func (e *ServerEngine) readTargetAddr(conn net.Conn) (string, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return "", err
	}

	length := binary.BigEndian.Uint32(lenBuf)
	encrypted := make([]byte, length)
	if _, err := io.ReadFull(conn, encrypted); err != nil {
		return "", err
	}

	decrypted, err := e.enc.Decrypt(encrypted)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func (e *ServerEngine) relay(client, target net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Target (decrypt)
	go func() {
		defer wg.Done()
		for {
			lenBuf := make([]byte, 4)
			if _, err := io.ReadFull(client, lenBuf); err != nil {
				return
			}

			length := binary.BigEndian.Uint32(lenBuf)
			encrypted := make([]byte, length)
			if _, err := io.ReadFull(client, encrypted); err != nil {
				return
			}

			decrypted, err := e.enc.Decrypt(encrypted)
			if err != nil {
				return
			}

			if _, err := target.Write(decrypted); err != nil {
				return
			}
		}
	}()

	// Target -> Client (encrypt)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := target.Read(buf)
			if err != nil {
				return
			}

			encrypted, err := e.enc.Encrypt(buf[:n])
			if err != nil {
				return
			}

			lenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)))

			if _, err := client.Write(lenBuf); err != nil {
				return
			}
			if _, err := client.Write(encrypted); err != nil {
				return
			}
		}
	}()

	wg.Wait()
}

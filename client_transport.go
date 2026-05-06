package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type ClientTransport struct {
	cfg        *ClientConfig
	enc        *ClientQuantumEncryption
	quicConn   quic.Connection
	httpsConn  net.Conn
	vncConn    net.Conn
}

func NewClientTransport(cfg *ClientConfig, enc *ClientQuantumEncryption) (*ClientTransport, error) {
	return &ClientTransport{
		cfg: cfg,
		enc: enc,
	}, nil
}

func (t *ClientTransport) SendQUIC(data []byte) error {
	if t.quicConn == nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"h3"},
		}
		
		addr := fmt.Sprintf("%s:%d", t.cfg.Server.Address, t.cfg.Server.Port)
		conn, err := quic.DialAddr(context.Background(), addr, tlsConfig, nil)
		if err != nil {
			return err
		}
		t.quicConn = conn
	}

	stream, err := t.quicConn.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}
	defer stream.Close()

	stream.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = stream.Write(data)
	return err
}

func (t *ClientTransport) SendHTTPS(data []byte) error {
	if t.httpsConn == nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		
		addr := fmt.Sprintf("%s:%d", t.cfg.Server.Address, t.cfg.Server.Port)
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		t.httpsConn = conn
	}

	t.httpsConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := t.httpsConn.Write(data)
	return err
}

func (t *ClientTransport) SendVNC(data []byte) error {
	if t.vncConn == nil {
		addr := fmt.Sprintf("%s:%d", t.cfg.Server.Address, t.cfg.Server.Port)
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			return err
		}
		t.vncConn = conn
	}

	t.vncConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := t.vncConn.Write(data)
	return err
}

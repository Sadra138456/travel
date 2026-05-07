package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/quic-go/quic-go"
)

type ClientTransport struct {
	serverAddr string
	tlsConfig  *tls.Config
	conn       *quic.Connection
}

func NewClientTransport(serverAddr string) *ClientTransport {
	return &ClientTransport{
		serverAddr: serverAddr,
		tlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"ant-protocol"},
		},
	}
}

func (t *ClientTransport) Connect(ctx context.Context) error {
	conn, err := quic.DialAddr(ctx, t.serverAddr, t.tlsConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to dial QUIC: %w", err)
	}
	t.conn = &conn
	return nil
}

func (t *ClientTransport) OpenStream(ctx context.Context) (net.Conn, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	stream, err := (*t.conn).OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &quicStreamConn{
		stream: stream,
		conn:   t.conn,
	}, nil
}

func (t *ClientTransport) Close() error {
	if t.conn != nil {
		return (*t.conn).CloseWithError(0, "client closing")
	}
	return nil
}

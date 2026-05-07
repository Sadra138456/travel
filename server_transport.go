package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type ServerTransport struct {
	listener *quic.Listener
}

func NewServerTransport(addr string) (*ServerTransport, error) {
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	listener, err := quic.ListenAddr(addr, tlsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &ServerTransport{
		listener: listener,
	}, nil
}

func (t *ServerTransport) Accept(ctx context.Context) (net.Conn, error) {
	conn, err := t.listener.Accept(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}

	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to accept stream: %w", err)
	}

	quicConn := conn
	return &quicStreamConn{
		stream: stream,
		conn:   &quicConn,
	}, nil
}

func (t *ServerTransport) Close() error {
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Ant Project"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certDER},
				PrivateKey:  key,
			},
		},
		NextProtos: []string{"ant-protocol"},
	}, nil
}

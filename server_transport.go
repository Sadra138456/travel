package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type ServerTransport struct {
	cfg          *ServerConfig
	enc          *ServerQuantumEncryption
	quicListener *quic.Listener
	httpsListener net.Listener
	vncListener   net.Listener
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewServerTransport(cfg *ServerConfig, enc *ServerQuantumEncryption) (*ServerTransport, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServerTransport{
		cfg:    cfg,
		enc:    enc,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (t *ServerTransport) StartQUIC(handler func(net.Conn, string)) error {
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		return err
	}

	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", t.cfg.ListenAddr, t.cfg.QUICPort)
	listener, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return err
	}

	t.quicListener = listener

	go func() {
		for {
			conn, err := listener.Accept(t.ctx)
			if err != nil {
				return
			}

			go func(c quic.Connection) {
				stream, err := c.AcceptStream(t.ctx)
				if err != nil {
					return
				}
				handler(&quicStreamConn{stream, c}, "QUIC")
			}(conn)
		}
	}()

	return nil
}

func (t *ServerTransport) StartHTTPS(handler func(net.Conn, string)) error {
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", t.cfg.ListenAddr, t.cfg.HTTPSPort)
	listener, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}

	t.httpsListener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handler(conn, "HTTPS")
		}
	}()

	return nil
}

func (t *ServerTransport) StartVNC(handler func(net.Conn, string)) error {
	addr := fmt.Sprintf("%s:%d", t.cfg.ListenAddr, t.cfg.VNCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.vncListener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			conn.Write([]byte("RFB 003.008\n"))

			go handler(conn, "VNC")
		}
	}()

	return nil
}

func (t *ServerTransport) Stop() {
	t.cancel()
	if t.quicListener != nil {
		t.quicListener.Close()
	}
	if t.httpsListener != nil {
		t.httpsListener.Close()
	}
	if t.vncListener != nil {
		t.vncListener.Close()
	}
}

type quicStreamConn struct {
	quic.Stream
	conn quic.Connection
}

func (q *quicStreamConn) LocalAddr() net.Addr {
	return q.conn.LocalAddr()
}

func (q *quicStreamConn) RemoteAddr() net.Addr {
	return q.conn.RemoteAddr()
}

func (q *quicStreamConn) SetDeadline(t time.Time) error {
	q.SetReadDeadline(t)
	q.SetWriteDeadline(t)
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
			Organization: []string{"Example Corp"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.

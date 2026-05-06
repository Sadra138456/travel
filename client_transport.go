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
	cfg  *ClientConfig
	enc  *ClientQuantumEncryption
	conn net.Conn
	ctx  context.Context
}

func NewClientTransport(cfg *ClientConfig, enc *ClientQuantumEncryption) *ClientTransport {
	return &ClientTransport{
		cfg: cfg,
		enc: enc,
		ctx: context.Background(),
	}
}

func (t *ClientTransport) ConnectQUIC() error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	addr := fmt.Sprintf("%s:%d", t.cfg.ServerAddr, t.cfg.QUICPort)
	conn, err := quic.DialAddr(t.ctx, addr, tlsConfig, nil)
	if err != nil {
		return err
	}

	stream, err := conn.OpenStreamSync(t.ctx)
	if err != nil {
		return err
	}

	t.conn = &quicStreamConn{stream, conn}
	return nil
}

func (t *ClientTransport) ConnectHTTPS() error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	addr := fmt.Sprintf("%s:%d", t.cfg.ServerAddr, t.cfg.HTTPSPort)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}

	t.conn = conn
	return nil
}

func (t *ClientTransport) ConnectVNC() error {
	addr := fmt.Sprintf("%s:%d", t.cfg.ServerAddr, t.cfg.VNCPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	buf := make([]byte, 12)
	conn.Read(buf)

	t.conn = conn
	return nil
}

func (t *ClientTransport) Write(data []byte) (int, error) {
	if t.conn == nil {
		return 0, fmt.Errorf("not connected")
	}
	encrypted := t.enc.Encrypt(data)
	return t.conn.Write(encrypted)
}

func (t *ClientTransport) Read(buf []byte) (int, error) {
	if t.conn == nil {
		return 0, fmt.Errorf("not connected")
	}
	n, err := t.conn.Read(buf)
	if err != nil {
		return n, err
	}
	decrypted := t.enc.Decrypt(buf[:n])
	copy(buf, decrypted)
	return len(decrypted), nil
}

func (t *ClientTransport) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
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

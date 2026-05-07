package main

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// quicStreamConn wraps a QUIC stream to implement net.Conn
type quicStreamConn struct {
	stream quic.Stream
	conn   *quic.Connection
}

func (q *quicStreamConn) Read(b []byte) (int, error) {
	return q.stream.Read(b)
}

func (q *quicStreamConn) Write(b []byte) (int, error) {
	return q.stream.Write(b)
}

func (q *quicStreamConn) Close() error {
	return q.stream.Close()
}

func (q *quicStreamConn) LocalAddr() net.Addr {
	return (*q.conn).LocalAddr()
}

func (q *quicStreamConn) RemoteAddr() net.Addr {
	return (*q.conn).RemoteAddr()
}

func (q *quicStreamConn) SetDeadline(t time.Time) error {
	return q.stream.SetDeadline(t)
}

func (q *quicStreamConn) SetReadDeadline(t time.Time) error {
	return q.stream.SetReadDeadline(t)
}

func (q *quicStreamConn) SetWriteDeadline(t time.Time) error {
	return q.stream.SetWriteDeadline(t)
}

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type ServerEngine struct {
	cfg       *ServerConfig
	enc       *ServerQuantumEncryption
	transport *ServerTransport
	sessions  map[string]*Session
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type Session struct {
	ID        string
	ClientIP  string
	CreatedAt time.Time
	LastSeen  time.Time
	BytesIn   uint64
	BytesOut  uint64
	Conn      net.Conn
}

func NewServerEngine(cfg *ServerConfig, enc *ServerQuantumEncryption) (*ServerEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	transport, err := NewServerTransport(cfg, enc)
	if err != nil {
		cancel()
		return nil, err
	}

	return &ServerEngine{
		cfg:       cfg,
		enc:       enc,
		transport: transport,
		sessions:  make(map[string]*Session),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (e *ServerEngine) Start() error {
	if err := e.transport.StartQUIC(e.handleConnection); err != nil {
		return fmt.Errorf("QUIC listener failed: %w", err)
	}

	if err := e.transport.StartHTTPS(e.handleConnection); err != nil {
		return fmt.Errorf("HTTPS listener failed: %w", err)
	}

	if err := e.transport.StartVNC(e.handleConnection); err != nil {
		return fmt.Errorf("VNC listener failed: %w", err)
	}

	go e.cleanupSessions()

	return nil
}

func (e *ServerEngine) Stop() {
	e.cancel()
	e.transport.Stop()
	
	e.mu.Lock()
	for _, sess := range e.sessions {
		if sess.Conn != nil {
			sess.Conn.Close()
		}
	}
	e.mu.Unlock()
}

func (e *ServerEngine) handleConnection(conn net.Conn, protocol string) {
	defer conn.Close()

	clientIP := conn.RemoteAddr().String()
	sessionID := fmt.Sprintf("%s-%d", clientIP, time.Now().UnixNano())

	session := &Session{
		ID:        sessionID,
		ClientIP:  clientIP,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		Conn:      conn,
	}

	e.mu.Lock()
	if len(e.sessions) >= e.cfg.MaxClients {
		e.mu.Unlock()
		log.Printf("Max clients reached, rejecting %s", clientIP)
		return
	}
	e.sessions[sessionID] = session
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.sessions, sessionID)
		e.mu.Unlock()
	}()

	log.Printf("[%s] New connection from %s via %s", sessionID[:8], clientIP, protocol)

	buf := make([]byte, 65535)
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Read error: %v", sessionID[:8], err)
			}
			return
		}

		session.LastSeen = time.Now()
		session.BytesIn += uint64(n)

		plaintext, err := e.enc.Decrypt(buf[:n])
		if err != nil {
			log.Printf("[%s] Decryption failed: %v", sessionID[:8], err)
			continue
		}

		response, err := e.forwardRequest(plaintext)
		if err != nil {
			log.Printf("[%s] Forward failed: %v", sessionID[:8], err)
			continue
		}

		encrypted, err := e.enc.Encrypt(response)
		if err != nil {
			log.Printf("[%s] Encryption failed: %v", sessionID[:8], err)
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		written, err := conn.Write(encrypted)
		if err != nil {
			log.Printf("[%s] Write error: %v", sessionID[:8], err)
			return
		}

		session.BytesOut += uint64(written)
	}
}

func (e *ServerEngine) forwardRequest(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("packet too short")
	}

	destLen := int(data[0])<<8 | int(data[1])
	if len(data) < 2+destLen {
		return nil, fmt.Errorf("invalid destination length")
	}

	destination := string(data[2 : 2+destLen])
	payload := data[2+destLen:]

	conn, err := net.DialTimeout("tcp", destination, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write(payload); err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	response := make([]byte, 65535)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	n, err := conn.Read(response)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return response[:n], nil
}

func (e *ServerEngine) cleanupSessions() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			now := time.Now()
			for id, sess := range e.sessions {
				if now.Sub(sess.LastSeen) > 5*time.Minute {
					log.Printf("[%s] Session timeout", id[:8])
					if sess.Conn != nil {
						sess.Conn.Close()
					}
					delete(e.sessions, id)
				}
			}
			e.mu.Unlock()
		}
	}
}

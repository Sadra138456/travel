package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Nanobot struct {
	ID        string
	Protocol  string
	transport *ClientTransport
	enc       *ClientQuantumEncryption
	active    bool
}

type NanobotPool struct {
	pool chan *Nanobot
	mu   sync.Mutex
}

func NewNanobotPool(size int) *NanobotPool {
	return &NanobotPool{
		pool: make(chan *Nanobot, size),
	}
}

func (p *NanobotPool) Get(protocol string, transport *ClientTransport, enc *ClientQuantumEncryption) *Nanobot {
	select {
	case bot := <-p.pool:
		bot.Protocol = protocol
		bot.active = true
		return bot
	default:
		return &Nanobot{
			ID:        generateID(),
			Protocol:  protocol,
			transport: transport,
			enc:       enc,
			active:    true,
		}
	}
}

func (p *NanobotPool) Put(bot *Nanobot) {
	bot.active = false
	select {
	case p.pool <- bot:
	default:
	}
}

func (b *Nanobot) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !b.active {
				return
			}
			b.sendPacket()
		}
	}
}

func (b *Nanobot) sendPacket() {
	size := 800 + rand.Intn(600)
	payload := make([]byte, size)
	rand.Read(payload)

	encrypted, _ := b.enc.Encrypt(payload)
	
	switch b.Protocol {
	case "QUIC":
		b.transport.SendQUIC(encrypted)
	case "HTTPS":
		b.transport.SendHTTPS(encrypted)
	case "VNC":
		b.transport.SendVNC(encrypted)
	}
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

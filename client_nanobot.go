package main

import (
	"crypto/rand"
	"sync"
	"time"
)

type Nanobot struct {
	ID        int
	Jitter    time.Duration
	PacketSize int
	Active    bool
}

type NanobotPool struct {
	bots []*Nanobot
	mu   sync.Mutex
}

func NewNanobotPool(count int) *NanobotPool {
	pool := &NanobotPool{
		bots: make([]*Nanobot, count),
	}

	for i := 0; i < count; i++ {
		pool.bots[i] = &Nanobot{
			ID:         i,
			Jitter:     time.Duration(50+i*10) * time.Millisecond,
			PacketSize: 512 + i*128,
			Active:     true,
		}
	}

	return pool
}

func (p *NanobotPool) Acquire() *Nanobot {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, bot := range p.bots {
		if bot.Active {
			return bot
		}
	}
	return p.bots[0]
}

func (p *NanobotPool) Release(bot *Nanobot) {
	// نیازی به کاری نیست - فقط برای سازگاری
}

func (p *NanobotPool) GetAll() []*Nanobot {
	return p.bots
}

func (n *Nanobot) GenerateTraffic() []byte {
	data := make([]byte, n.PacketSize)
	rand.Read(data)
	return data
}

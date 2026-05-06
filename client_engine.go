package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type ClientEngine struct {
	cfg       *ClientConfig
	enc       *ClientQuantumEncryption
	transport *ClientTransport
	nanobots  []*Nanobot
	pool      *NanobotPool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	stats     struct {
		BytesSent uint64
		BytesRecv uint64
		Packets   uint64
	}
}

func NewClientEngine(cfg *ClientConfig, enc *ClientQuantumEncryption) (*ClientEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	transport, err := NewClientTransport(cfg, enc)
	if err != nil {
		cancel()
		return nil, err
	}

	pool := NewNanobotPool(cfg.Stealth.MaxBots)

	return &ClientEngine{
		cfg:       cfg,
		enc:       enc,
		transport: transport,
		pool:      pool,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (e *ClientEngine) Start() error {
	initialBots := e.cfg.Stealth.MinBots
	for i := 0; i < initialBots; i++ {
		protocol := e.selectProtocol()
		bot := e.pool.Get(protocol, e.transport, e.enc)
		e.nanobots = append(e.nanobots, bot)
		go bot.Run(e.ctx)
	}

	go e.trafficShaper()
	go e.dynamicScaling()

	return nil
}

func (e *ClientEngine) Stop() {
	e.cancel()
	e.mu.Lock()
	for _, bot := range e.nanobots {
		e.pool.Put(bot)
	}
	e.nanobots = nil
	e.mu.Unlock()
}

func (e *ClientEngine) selectProtocol() string {
	r := rand.Float64()
	if r < e.cfg.Stealth.QUICRatio {
		return "QUIC"
	} else if r < e.cfg.Stealth.QUICRatio+e.cfg.Stealth.HTTPSRatio {
		return "HTTPS"
	}
	return "VNC"
}

func (e *ClientEngine) trafficShaper() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
		}
	}
}

func (e *ClientEngine) dynamicScaling() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			current := len(e.nanobots)
			target := e.cfg.Stealth.MinBots + rand.Intn(10)
			
			if target > current {
				for i := 0; i < target-current; i++ {
					protocol := e.selectProtocol()
					bot := e.pool.Get(protocol, e.transport, e.enc)
					e.nanobots = append(e.nanobots, bot)
					go bot.Run(e.ctx)
				}
			} else if target < current {
				for i := 0; i < current-target; i++ {
					if len(e.nanobots) > 0 {
						bot := e.nanobots[len(e.nanobots)-1]
						e.nanobots = e.nanobots[:len(e.nanobots)-1]
						e.pool.Put(bot)
					}
				}
			}
			e.mu.Unlock()
		}
	}
}

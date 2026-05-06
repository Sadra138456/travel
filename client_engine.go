package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type ClientEngine struct {
	config    ClientConfig
	quantum   *ClientQuantumEncryption
	transport *ClientTransport
	nanobots  *NanobotPool
	mu        sync.RWMutex
	running   bool
}

func NewClientEngine(config ClientConfig, quantum *ClientQuantumEncryption) *ClientEngine {
	return &ClientEngine{
		config:  config,
		quantum: quantum,
	}
}

func (e *ClientEngine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("client already running")
	}

	// ایجاد transport layer
	e.transport = NewClientTransport(e.config, e.quantum)

	// ایجاد nanobot pool
	e.nanobots = NewNanobotPool(e.config.Stealth.MinBots, e.config.Stealth.MaxBots)

	e.running = true

	// شروع ارسال ترافیک
	go e.trafficLoop()

	log.Println("✅ Client engine started successfully")
	return nil
}

func (e *ClientEngine) trafficLoop() {
	ticker := time.NewTicker(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C

		if !e.running {
			break
		}

		// انتخاب تصادفی پروتکل
		protocol := e.selectProtocol()

		// ارسال داده تصادفی
		data := e.generateRandomData()

		// ارسال از طریق nanobot
		bot := e.nanobots.Acquire()
		go func(b *Nanobot, proto string, payload []byte) {
			defer e.nanobots.Release(b)

			if err := e.transport.Send(proto, payload); err != nil {
				log.Printf("⚠️  Send failed via %s: %v", proto, err)
			} else {
				log.Printf("📤 Sent %d bytes via %s (bot #%d)", len(payload), proto, b.ID)
			}
		}(bot, protocol, data)

		// تنظیم تایمر بعدی
		ticker.Reset(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
	}
}

func (e *ClientEngine) selectProtocol() string {
	r := rand.Float64()

	if r < e.config.Stealth.QUICRatio {
		return "quic"
	} else if r < e.config.Stealth.QUICRatio+e.config.Stealth.HTTPSRatio {
		return "https"
	}
	return "vnc"
}

func (e *ClientEngine) generateRandomData() []byte {
	size := 800 + rand.Intn(600) // 800-1400 bytes
	data := make([]byte, size)
	rand.Read(data)
	return data
}

func (e *ClientEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.running = false
	log.Println("🛑 Client engine stopped")
}

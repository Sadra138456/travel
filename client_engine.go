package main

import (
	"log"
	"time"
)

type ClientEngine struct {
	cfg       *ClientConfig
	transport *ClientTransport
	nanobots  *NanobotPool
	enc       *ClientQuantumEncryption
}

func NewClientEngine(cfg *ClientConfig) (*ClientEngine, error) {
	enc, err := NewClientQuantumEncryption(cfg.Password, cfg.Salt)
	if err != nil {
		return nil, err
	}

	return &ClientEngine{
		cfg: cfg,
		enc: enc,
	}, nil
}

func (e *ClientEngine) Start() error {
	e.transport = NewClientTransport(e.cfg, e.enc)
	e.nanobots = NewNanobotPool(e.cfg.NanobotCount)

	var err error
	switch e.cfg.Protocol {
	case "QUIC":
		err = e.transport.ConnectQUIC()
	case "HTTPS":
		err = e.transport.ConnectHTTPS()
	case "VNC":
		err = e.transport.ConnectVNC()
	}

	if err != nil {
		return err
	}

	log.Printf("✓ Connected via %s", e.cfg.Protocol)

	go e.trafficLoop()

	return nil
}

func (e *ClientEngine) trafficLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		bot := e.nanobots.Acquire()
		if bot == nil {
			continue
		}

		data := bot.GenerateTraffic()
		_, err := e.transport.Write(data)
		if err != nil {
			log.Printf("Send error: %v", err)
		}

		e.nanobots.Release(bot)
		time.Sleep(bot.Jitter)
	}
}

func (e *ClientEngine) Stop() {
	if e.transport != nil {
		e.transport.Close()
	}
}

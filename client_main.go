package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type ClientConfig struct {
	Server struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
	} `json:"server"`
	Encryption struct {
		Password string `json:"password"`
		Salt     string `json:"salt"`
	} `json:"encryption"`
	Stealth struct {
		QUICRatio  float64 `json:"quic_ratio"`
		HTTPSRatio float64 `json:"https_ratio"`
		VNCRatio   float64 `json:"vnc_ratio"`
		MaxBurstKB int     `json:"max_burst_kb"`
		MinBots    int     `json:"min_bots"`
		MaxBots    int     `json:"max_bots"`
	} `json:"stealth"`
}

func runClient() {
	configFile := os.Args[1]
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	quantum, err := NewClientQuantumEncryption(config.Encryption.Password, config.Encryption.Salt)
	if err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	engine := NewClientEngine(config, quantum)

	log.Printf("🚀 ANT Client started")
	log.Printf("📡 Connecting to %s:%d", config.Server.Address, config.Server.Port)
	log.Printf("🔐 Quantum-Resistant Encryption: Enabled")
	log.Printf("🤖 Dynamic Nanobots: %d-%d", config.Stealth.MinBots, config.Stealth.MaxBots)
	log.Printf("🎭 Protocol Mix: QUIC(%.0f%%) HTTPS(%.0f%%) VNC(%.0f%%)",
		config.Stealth.QUICRatio*100,
		config.Stealth.HTTPSRatio*100,
		config.Stealth.VNCRatio*100)

	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	// نگه‌داشتن برنامه
	select {}
}

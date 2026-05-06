package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
		QUICRatio   float64 `json:"quic_ratio"`
		HTTPSRatio  float64 `json:"https_ratio"`
		VNCRatio    float64 `json:"vnc_ratio"`
		MaxBurstKB  int     `json:"max_burst_kb"`
		MinBots     int     `json:"min_bots"`
		MaxBots     int     `json:"max_bots"`
	} `json:"stealth"`
	LocalSOCKS struct {
		Enabled bool   `json:"enabled"`
		Port    int    `json:"port"`
	} `json:"local_socks"`
}

func runClient() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./ant-project client_config.json")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var cfg ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	enc, err := NewClientQuantumEncryption(cfg.Encryption.Password, cfg.Encryption.Salt)
	if err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	engine, err := NewClientEngine(&cfg, enc)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	fmt.Printf("✓ ANT Client connected to %s:%d\n", cfg.Server.Address, cfg.Server.Port)
	if cfg.LocalSOCKS.Enabled {
		fmt.Printf("  SOCKS5 proxy: 127.0.0.1:%d\n", cfg.LocalSOCKS.Port)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	engine.Stop()
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ServerConfig struct {
	ListenAddr    string `json:"listen_addr"`
	QUICPort      int    `json:"quic_port"`
	HTTPSPort     int    `json:"https_port"`
	VNCPort       int    `json:"vnc_port"`
	Password      string `json:"password"`
	Salt          string `json:"salt"`
	MaxClients    int    `json:"max_clients"`
}

func runServer() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./ant-project server_config.json")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	enc, err := NewServerQuantumEncryption(cfg.Password, cfg.Salt)
	if err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	engine, err := NewServerEngine(&cfg, enc)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Printf("✓ ANT Server started\n")
	fmt.Printf("  QUIC:  %s:%d\n", cfg.ListenAddr, cfg.QUICPort)
	fmt.Printf("  HTTPS: %s:%d\n", cfg.ListenAddr, cfg.HTTPSPort)
	fmt.Printf("  VNC:   %s:%d\n", cfg.ListenAddr, cfg.VNCPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	engine.Stop()
}

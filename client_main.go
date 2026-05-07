package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type ClientConfig struct {
	ServerAddr string `json:"server_addr"`
	Password   string `json:"password"`
	Salt       string `json:"salt"`
	LocalAddr  string `json:"local_addr"`
}

func runClient(configPath string) error {
	// Read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Create encryption
	enc := NewClientQuantumEncryption(config.Password, config.Salt)

	// Create transport
	transport := NewClientTransport(config.ServerAddr)

	// Create engine
	engine := NewClientEngine(transport, enc, config.LocalAddr)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down client...")
		cancel()
	}()

	// Start engine
	fmt.Printf("Starting client, listening on %s\n", config.LocalAddr)
	return engine.Start(ctx)
}

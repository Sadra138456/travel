package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type ServerConfig struct {
	ListenAddr string `json:"listen_addr"`
	Password   string `json:"password"`
	Salt       string `json:"salt"`
}

func runServer(configPath string) error {
	// Read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Create encryption
	enc := NewServerQuantumEncryption(config.Password, config.Salt)

	// Create transport
	transport, err := NewServerTransport(config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}
	defer transport.Close()

	// Create engine
	engine := NewServerEngine(transport, enc)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	// Start engine
	fmt.Printf("Server listening on %s\n", config.ListenAddr)
	return engine.Start(ctx)
}

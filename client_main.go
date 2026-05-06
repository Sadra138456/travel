package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ClientConfig struct {
	ServerAddr   string `json:"server_addr"`
	QUICPort     int    `json:"quic_port"`
	HTTPSPort    int    `json:"https_port"`
	VNCPort      int    `json:"vnc_port"`
	Protocol     string `json:"protocol"`
	Password     string `json:"password"`
	Salt         string `json:"salt"`
	NanobotCount int    `json:"nanobot_count"`
}

func loadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func runClient(configPath string) error {
	cfg, err := loadClientConfig(configPath)
	if err != nil {
		return err
	}

	engine, err := NewClientEngine(cfg)
	if err != nil {
		return err
	}

	if err := engine.Start(); err != nil {
		return err
	}

	log.Println("✓ Client started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	engine.Stop()
	log.Println("✓ Client stopped")

	return nil
}

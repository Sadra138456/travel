package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: travel <config.json>")
		os.Exit(1)
	}

	configPath := os.Args[1]

	// Determine if this is client or server based on config filename
	if strings.Contains(configPath, "client") {
		if err := runClient(configPath); err != nil {
			fmt.Printf("Client error: %v\n", err)
			os.Exit(1)
		}
	} else if strings.Contains(configPath, "server") {
		if err := runServer(configPath); err != nil {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Config file must contain 'client' or 'server' in its name")
		os.Exit(1)
	}
}

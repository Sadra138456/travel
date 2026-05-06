package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: travel <config.json>")
	}

	configPath := os.Args[1]

	// تشخیص کلاینت یا سرور بر اساس نام فایل
	if configPath == "client_config.json" {
		if err := runClient(configPath); err != nil {
			log.Fatal(err)
		}
	} else if configPath == "server_config.json" {
		if err := runServer(configPath); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Config file must be 'client_config.json' or 'server_config.json'")
	}
}

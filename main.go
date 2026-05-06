package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  Client: go run . client_config.json")
		fmt.Println("  Server: go run . server_config.json")
		os.Exit(1)
	}

	configFile := os.Args[1]
	
	if isClientConfig(configFile) {
		runClient()
	} else {
		runServer()
	}
}

func isClientConfig(filename string) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	
	return len(data) > 0 && (string(data)[0:20] == `{
  "server": {
    "a` || containsString(string(data), `"server":`))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

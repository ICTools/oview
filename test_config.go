// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/yourusername/oview/internal/config"
)

func main() {
	// Test global config
	cfg := config.DefaultGlobalConfig()
	fmt.Println("Default global config:")
	fmt.Printf("  Postgres: %s:%d (user: %s)\n", cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser)
	fmt.Printf("  n8n: %s\n", cfg.N8nURL)
	fmt.Printf("  Network: %s\n", cfg.DockerNetworkName)

	// Test validation
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ Validation passed")

	// Test DSN generation
	dsn := cfg.GetDSN("testdb")
	fmt.Printf("  DSN: %s\n", dsn)

	fmt.Println("\n✓ Config package working correctly")
}

package main

import (
	"log"
	"os"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start the server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

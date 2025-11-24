package main

import (
	"log"

	"github.com/abikandiah/task-worker/internal/factory"
)

// Entry point for HTTP server
func main() {
	deps, err := factory.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
}

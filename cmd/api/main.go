package main

import (
	"log"

	"github.com/abikandiah/task-worker/internal/container"
	"github.com/abikandiah/task-worker/internal/platform/executor"
)

// Entry point for HTTP server
func main() {
	deps, err := container.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("API server startup failed: %v", err)
	}

	taskFactory := executor.InitTaskFactory(deps)
}

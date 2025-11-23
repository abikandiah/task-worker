package main

import (
	"log"

	"github.com/abikandiah/task-worker/internal/factory"
	"github.com/abikandiah/task-worker/internal/service"
)

// Entry point for HTTP server
func main() {
	deps, err := factory.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	taskFactory := factory.InitTaskFactory(deps)
	jobService := service.NewJobService(deps, taskFactory)
}

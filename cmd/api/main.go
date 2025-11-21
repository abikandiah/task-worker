package main

import (
	"log"

	"github.com/abikandiah/task-worker/internal/container"
	"github.com/abikandiah/task-worker/internal/service"
)

// Entry point for HTTP server
func main() {
	deps, err := container.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("API server startup failed: %v", err)
	}

	taskFactory := container.InitTaskFactory(deps)

	taskExecutor := &service.TaskExecutor{
		Logger:      deps.Logger,
		TaskFactory: *taskFactory,
	}
}

package main

import (
	"log"

	"github.com/abikandiah/task-worker/internal/container"
	"github.com/abikandiah/task-worker/internal/core/port"
	"github.com/abikandiah/task-worker/internal/task"
)

// Entry point for HTTP server
func main() {
	deps, err := container.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("API server startup failed: %v", err)
	}

	taskFactory := task.NewTaskFactory()
	taskFactory.Register("email_send_task", func(params any) (port.Task, error) {
		return task.NewEmailSendTask(params, &task.EmailDependencies{
			Logger: deps.Logger,
		})
	})
}

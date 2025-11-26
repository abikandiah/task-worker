package main

import (
	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/app"
	"github.com/abikandiah/task-worker/internal/mock"
)

// Entry point for HTTP server
func main() {
	app := app.NewApplication(&app.AppDependencies{
		Config:     config.MustLoad(),
		Repository: mock.NewMockRepo(),
	})

	app.Run()
}

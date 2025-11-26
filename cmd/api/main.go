package main

import (
	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/app"
)

// Entry point for HTTP server
func main() {
	app := app.NewApplication(&app.AppDependencies{
		Config: config.MustLoad(),
	})

	app.Run()
}

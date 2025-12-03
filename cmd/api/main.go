package main

import (
	"flag"

	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/app"
)

// Entry point for HTTP server
func main() {
	flag.Parse()

	app := app.NewApplication(&app.AppDependencies{
		Config: config.MustLoad(),
	})

	if *config.MigrateFlag {
		app.RunMigrations()
	}

	app.Run()
}

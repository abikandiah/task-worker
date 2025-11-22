package container

import (
	"log/slog"
	"os"
)

func setupLogger() *slog.Logger {
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)

	// Create a JSON handler that writes to standard output
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	})

	return slog.New(handler)
}

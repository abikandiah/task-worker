package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/abikandiah/task-worker/config"
)

func SetupLogger(cfg config.LoggerConfig) *slog.Logger {
	logLevel := parseLogLevel(cfg.Level)
	levelVar := new(slog.LevelVar)
	levelVar.Set(logLevel)

	// Determine output format based on environment
	var handler slog.Handler

	if cfg.Environment == "development" || cfg.Environment == "dev" {
		// Use text handler for better readability in development
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     levelVar,
			AddSource: true,
		})
	} else {
		// Use JSON handler for production (easier to parse and index)
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     levelVar,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Rename source key for better clarity
				if a.Key == slog.SourceKey {
					return slog.Attr{
						Key:   "caller",
						Value: a.Value,
					}
				}
				return a
			},
		})
	}

	// Create logger with default attributes
	logger := slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.Version),
		slog.String("environment", cfg.Environment),
	)

	// Set as default logger
	slog.SetDefault(logger)

	// Log initialization
	logger.Info("logger initialized",
		slog.String("level", cfg.Level),
		slog.String("environment", cfg.Environment),
		slog.String("format", getHandlerType(cfg)),
	)

	return logger
}

func getHandlerType(cfg config.LoggerConfig) string {
	if cfg.Environment == "development" || cfg.Environment == "dev" {
		return "text"
	}
	return "json"
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

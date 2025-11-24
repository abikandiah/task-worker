package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type LoggerParams struct {
	Level       string
	Environment string
	ServiceName string
	Version     string
}

func SetupLogger(config LoggerParams) *slog.Logger {
	logLevel := parseLogLevel(config.Level)
	levelVar := new(slog.LevelVar)
	levelVar.Set(logLevel)

	// Output format based on environment
	var handler slog.Handler

	if config.Environment == "development" || config.Environment == "dev" {
		// Use text handler for better readability in development
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     levelVar,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				a = shortenSource(groups, a)
				a = replaceEmptyAttr(groups, a)
				return a
			},
		})
	} else {
		// Use JSON handler for production (easier to parse and index)
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:       levelVar,
			AddSource:   false,
			ReplaceAttr: replaceEmptyAttr,
		})
	}

	// Create logger with default attributes
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log initialization
	logger.Info("logger initialized",
		slog.String("level", config.Level),
		slog.String("environment", config.Environment),
		slog.String("service", config.ServiceName),
		slog.String("version", config.Version),
		slog.String("format", getHandlerType(&config)),
	)

	return logger
}

func shortenSource(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.SourceKey {
		return a
	}

	source, ok := a.Value.Any().(*slog.Source)
	if !ok {
		return a
	}

	// Split path
	parts := strings.Split(source.File, string(os.PathSeparator))

	if len(parts) >= 2 {
		// Join "parent_dir/filename".
		source.File = filepath.Join(parts[len(parts)-2], parts[len(parts)-1])
	} else if len(parts) == 1 {
		source.File = parts[0]
	}

	// Note: We return the original attribute 'a', but its underlying
	// *slog.Source struct (a pointer) has been modified.
	return a
}

// Return empty attr to hide from log output
func replaceEmptyAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindString && a.Value.String() == "" {
		return slog.Attr{}
	}
	return a
}

func getHandlerType(config *LoggerParams) string {
	if config.Environment == "development" || config.Environment == "dev" {
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

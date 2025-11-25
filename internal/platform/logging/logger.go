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
	options := &slog.HandlerOptions{
		Level:       levelVar,
		ReplaceAttr: replaceAttrFunc,
	}

	var baseHandler slog.Handler

	if config.Environment == "development" || config.Environment == "dev" {
		options.AddSource = true
		baseHandler = slog.NewTextHandler(os.Stderr, options)
	} else {
		baseHandler = slog.NewJSONHandler(os.Stderr, options)
	}

	// Create logger with default attributes
	logger := slog.New(NewContextHandler(baseHandler))
	slog.SetDefault(logger)

	// Log initialization
	logger.Info("logger and config initialized",
		slog.String("environment", config.Environment),
		slog.String("service", config.ServiceName),
		slog.String("version", config.Version),
		slog.String("format", getHandlerType(&config)),
	)

	return logger
}

func replaceAttrFunc(groups []string, a slog.Attr) slog.Attr {
	// Shorten source
	if a.Key == slog.SourceKey {
		return shortenSource(groups, a)
	}
	// Format string output
	if a.Value.Kind() == slog.KindDuration {
		return replaceDuration(groups, a)
	}
	// Hide empty settings
	if a.Value.Kind() == slog.KindString && a.Value.String() == "" {
		return slog.Attr{}
	}
	// // Format for UTC
	// if a.Key == slog.TimeKey && len(groups) == 0 {
	// 	return replaceTimeAttr(groups, a)
	// }
	return a
}

func shortenSource(_ []string, a slog.Attr) slog.Attr {
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

func replaceDuration(_ []string, a slog.Attr) slog.Attr {
	duration := a.Value.Duration()
	// Return new Attr with string value
	return slog.String(a.Key, duration.String())
}

// func replaceTimeAttr(_ []string, a slog.Attr) slog.Attr {
// 	t := a.Value.Time()
// 	t = t.UTC()
// 	// 2. Explicitly format the UTC time as an ISO 8601 string with 'Z'
// 	// Note: The time.RFC3339Nano constant is generally preferred
// 	// and includes the 'Z' to denote UTC.
// 	a.Value = slog.StringValue(t.Format(time.RFC3339Nano))
// 	return a
// }

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

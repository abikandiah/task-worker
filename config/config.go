package config

import (
	"fmt"

	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/abikandiah/task-worker/internal/platform/server"
	"github.com/abikandiah/task-worker/internal/service"
)

// EnvPrefix is the mandatory prefix for all environment variables (e.g., APP_SERVER_PORT)
const EnvPrefix = "APP"

type Config struct {
	Environment string          `mapstructure:"environment"`
	ServiceName string          `mapstructure:"service_name"`
	Level       string          `mapstructure:"level"`
	Version     string          `mapstructure:"version"`
	Worker      *service.Config `mapstructure:"worker"`
	Server      *server.Config  `mapstructure:"server"`
	Database    *db.Config      `mapstructure:"database"`
}

// Loads configuration and panics if it fails
func MustLoad() *Config {
	config, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}

// Loads configuration from file and panics if it fails
func MustLoadFromFile(path string) *Config {
	config, err := LoadFromFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config from file: %v", err))
	}
	return config
}

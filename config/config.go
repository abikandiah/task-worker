package config

import (
	"flag"
	"fmt"

	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/abikandiah/task-worker/internal/platform/logging"
	"github.com/abikandiah/task-worker/internal/platform/server"
	"github.com/abikandiah/task-worker/internal/service"
)

// EnvPrefix is the mandatory prefix for all environment variables (e.g., APP_SERVER_PORT)
const EnvPrefix = "APP"

// Application flags
var (
	MigrateFlag *bool
	ConfigPath  *string
)

func init() {
	MigrateFlag = flag.Bool("migrate", false, "Run database migrations and exit.")
	ConfigPath = flag.String("config", "", "Path to the configuration file (e.g., config.yaml)")
}

type Config struct {
	Environment string          `mapstructure:"environment"`
	ServiceName string          `mapstructure:"service_name"`
	Level       string          `mapstructure:"level"`
	Version     string          `mapstructure:"version"`
	Worker      *service.Config `mapstructure:"worker"`
	Server      *server.Config  `mapstructure:"server"`
	Database    *db.Config      `mapstructure:"database"`
	Logger      *logging.Config `mapstructure:"logger"`
}

func (config *Config) updateLoggingEnvironment() {
	config.Logger.Environment = config.Environment
	config.Logger.ServiceName = config.ServiceName
	config.Logger.Version = config.Version
}

// Loads configuration and panics if it fails
func MustLoad() *Config {
	config, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	config.updateLoggingEnvironment()
	return config
}

// Loads configuration from file and panics if it fails
func MustLoadFromFile(path string) *Config {
	config, err := LoadFromFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config from file: %v", err))
	}
	config.updateLoggingEnvironment()
	return config
}

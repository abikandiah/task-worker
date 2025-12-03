package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/abikandiah/task-worker/internal/platform/logging"
	"github.com/abikandiah/task-worker/internal/platform/server"
	"github.com/abikandiah/task-worker/internal/service"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Load configuration with priority (highest to lowest):
// 1. Environment variables
// 2. Config file (config.yaml)
// 3. Default values
func Load() (*Config, error) {
	v := initDefaultViper()

	// Enable environment variables
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind specific env vars without prefix for common cases
	bindEnvironmentVariables(v)

	// Configure viper
	if *ConfigPath != "" {
		v.SetConfigFile(*ConfigPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")        // Look in current directory
		v.AddConfigPath("./config") // Look in config directory
		v.AddConfigPath("~/.config/task-worker")
	}

	// Read config file (optional - won't error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; using defaults and env vars
	}

	// Unmarshal config into struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// LoadFromFile reads configuration from a specific file path
func LoadFromFile(path string) (*Config, error) {
	v := initDefaultViper()

	// Set specific config file
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Enable environment variables (they override file values)
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind specific env vars
	bindEnvironmentVariables(v)

	// Unmarshal config into struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Default configuration
func setDefaults(v *viper.Viper) {
	v.SetDefault("environment", "development")
	v.SetDefault("service_name", "task-worker")
	v.SetDefault("version", "1.0.0")

	server.SetConfigDefaults(v)
	db.SetConfigDefaults(v)
	service.SetConfigDefaults(v)
	logging.SetConfigDefaults(v)
}

// bindEnvironmentVariables explicitly binds environment variables
// This allows using common env var names without APP_ prefix
func bindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("environment", "ENVIRONMENT")
	v.BindEnv("service_name", "SERVICE_NAME")
	v.BindEnv("version", "SERVICE_VERSION")

	server.BindEnvironmentVariables(v)
	db.BindEnvironmentVariables(v)
	service.BindEnvironmentVariables(v)
	logging.BindEnvironmentVariables(v)
}

func initDefaultViper() *viper.Viper {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	v := viper.New()
	setDefaults(v)
	return v
}

func (config *Config) validate() error {
	if err := config.Server.Validate(); err != nil {
		return err
	}
	if err := config.Database.Validate(); err != nil {
		return err
	}
	// Safety check for production password
	isProd := strings.ToLower(config.Environment) == "production"
	if config.Database.Password == "" && isProd {
		// Check if password was provided via environment variables
		if os.Getenv("DB_PASSWORD") == "" && os.Getenv(EnvPrefix+"_DATABASE_PASSWORD") == "" {
			return fmt.Errorf("database password cannot be empty in production environment")
		}
	}
	if err := config.Worker.Validate(); err != nil {
		return err
	}
	return nil
}

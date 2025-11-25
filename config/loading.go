package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")        // Look in current directory
	v.AddConfigPath("./config") // Look in config directory
	v.AddConfigPath("/etc/app") // Look in /etc/app

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

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)
	// --- CORS Configuration Defaults ---
	v.SetDefault("server.cors.enabled", false)
	v.SetDefault("server.cors.allowed_origins", []string{"*"})
	v.SetDefault("server.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("server.cors.allowed_headers", []string{"Content-Type", "Authorization"})
	v.SetDefault("server.cors.allow_credentials", false)
	v.SetDefault("server.cors.max_age", 1*time.Hour)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.db_name", "myapp")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("database.ssl_mode", "disable")

	v.SetDefault("worker.job_buffer_capacity", 128)
	v.SetDefault("worker.job_worker_count", 2)
	v.SetDefault("worker.task_worker_count", 4)

	v.SetDefault("logger.level", "INFO")
	v.SetDefault("rate_limit.requests_per_second", 100)
	v.SetDefault("rate_limit.burst", 200)
}

// bindEnvironmentVariables explicitly binds environment variables
// This allows using common env var names without APP_ prefix
func bindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("environment", "ENVIRONMENT")
	v.BindEnv("service_name", "SERVICE_NAME")
	v.BindEnv("version", "SERVICE_VERSION")

	// Server Config
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	v.BindEnv("server.timeout", "SERVER_TIMEOUT")
	v.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

	// Database Config
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.db_name", "DB_NAME")
	v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	v.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	// Worker Config
	v.BindEnv("worker.job_buffer_capacity", "JOB_BUFFER_CAPACITY")
	v.BindEnv("worker.job_worker_count", "JOB_WORKER_COUNT")
	v.BindEnv("worker.task_worker_count", "TASK_WORKER_COUNT")

	// Logger Config
	v.BindEnv("logger.level", "LOG_LEVEL")

	// Rate Limit Config
	v.BindEnv("rate_limit.requests_per_second", "RATE_LIMIT_REQUESTS_PER_SECOND")
	v.BindEnv("rate_limit.burst", "RATE_LIMIT_BURST")
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
	if config.Server.ReadTimeout <= 0 || config.Server.WriteTimeout <= 0 ||
		config.Server.IdleTimeout <= 0 || config.Server.ShutdownTimeout <= 0 ||
		config.Server.Timeout <= 0 {

		return fmt.Errorf("server timeouts must be positive")
	}
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Port < 1 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}
	if config.Database.MaxOpenConns < 1 {
		return fmt.Errorf("max open connections must be at least 1")
	}
	if config.Database.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	if config.Database.MaxIdleConns > config.Database.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}

	// Safety check for production password
	isProd := strings.ToLower(config.Environment) == "production"
	if config.Database.Password == "" && isProd {
		// Check if password was provided via environment variables
		if os.Getenv("DB_PASSWORD") == "" && os.Getenv(EnvPrefix+"_DATABASE_PASSWORD") == "" {
			return fmt.Errorf("database password cannot be empty in production environment")
		}
	}

	if config.Worker.JobBufferCapacity < 0 {
		return fmt.Errorf("job buffer capacity cannot be negative")
	}
	if config.Worker.JobWorkerCount < 1 {
		return fmt.Errorf("job worker count must be at least 1")
	}
	if config.Worker.TaskWorkerCount < 1 {
		return fmt.Errorf("task worker count must be at least 1")
	}

	if config.RateLimit.Burst < 1 || config.RateLimit.RequestsPerSecond < 1 {
		return fmt.Errorf("rate limits must be greater than 0")
	}

	return nil
}

package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// EnvPrefix is the mandatory prefix for all environment variables (e.g., APP_SERVER_PORT)
const EnvPrefix = "APP"

type Config struct {
	Worker   WorkerConfig   `mapstructure:"worker"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	SSLMode         string        `mapstructure:"ssl_mode"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Environment string `mapstructure:"environment"`
	ServiceName string `mapstructure:"service_name"`
	Version     string `mapstructure:"version"`
}

type WorkerConfig struct {
	JobBufferCapacity int `mapstructure:"job_buffer_capacity"`
	JobWorkerCount    int `mapstructure:"job_worker_count"`
	TaskWorkerCount   int `mapstructure:"task_worker_count"`
}

// Load configuration with priority (highest to lowest):
// 1. Environment variables
// 2. Config file (config.yaml)
// 3. Default values
func Load() (*Config, error) {
	v := viper.New()

	setDefaults(v)

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

	// Enable environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific env vars without prefix for common cases
	bindEnvironmentVariables(v)

	// Unmarshal config into struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// LoadFromFile reads configuration from a specific file path
func LoadFromFile(path string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set specific config file
	v.SetConfigFile(path)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Enable environment variables (they override file values)
	v.AutomaticEnv()
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific env vars
	bindEnvironmentVariables(v)

	// Unmarshal config into struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Default configuration
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.db_name", "myapp")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("database.ssl_mode", "disable")

	v.SetDefault("logger.level", "INFO")
	v.SetDefault("logger.environment", "development")
	v.SetDefault("logger.service_name", "task-worker")
	v.SetDefault("logger.version", "1.0.0")

	v.SetDefault("worker.job_buffer_capacity", 128)
	v.SetDefault("worker.job_worker_count", 2)
	v.SetDefault("worker.task_worker_count", 4)
}

// bindEnvironmentVariables explicitly binds environment variables
// This allows using common env var names without APP_ prefix
func bindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.db_name", "DB_NAME")
	v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	v.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	v.BindEnv("logger.level", "LOG_LEVEL")
	v.BindEnv("logger.environment", "ENVIRONMENT")
	v.BindEnv("logger.service_name", "SERVICE_NAME")
	v.BindEnv("logger.version", "SERVICE_VERSION")

	v.BindEnv("worker.job_buffer_capacity", "JOB_BUFFER_CAPACITY")
	v.BindEnv("worker.job_worker_count", "JOB_WORKER_COUNT")
	v.BindEnv("worker.task_worker_count", "TASK_WORKER_COUNT")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.ReadTimeout <= 0 || c.Server.WriteTimeout <= 0 || c.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("server timeouts must be positive")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Database.MaxOpenConns < 1 {
		return fmt.Errorf("max open connections must be at least 1")
	}
	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}

	// Safety check for production password
	isProd := strings.ToLower(c.Logger.Environment) == "production"
	if c.Database.Password == "" && isProd {
		// Check if password was provided via environment variables
		if os.Getenv("DB_PASSWORD") == "" && os.Getenv(EnvPrefix+"_DATABASE_PASSWORD") == "" {
			return fmt.Errorf("database password cannot be empty in production environment")
		}
	}

	if c.Worker.JobBufferCapacity < 0 {
		return fmt.Errorf("job buffer capacity cannot be negative")
	}
	if c.Worker.JobWorkerCount < 1 {
		return fmt.Errorf("job worker count must be at least 1")
	}
	if c.Worker.TaskWorkerCount < 1 {
		return fmt.Errorf("task worker count must be at least 1")
	}

	return nil
}

// GetDatabaseDSN returns a PostgreSQL DSN connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Loads configuration and panics if it fails
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// Loads configuration from file and panics if it fails
func MustLoadFromFile(path string) *Config {
	cfg, err := LoadFromFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config from file: %v", err))
	}
	return cfg
}

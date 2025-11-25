package config

import (
	"fmt"
	"log/slog"
	"time"
)

// EnvPrefix is the mandatory prefix for all environment variables (e.g., APP_SERVER_PORT)
const EnvPrefix = "APP"

type Config struct {
	Environment string           `mapstructure:"environment"`
	ServiceName string           `mapstructure:"service_name"`
	Version     string           `mapstructure:"version"`
	Worker      *WorkerConfig    `mapstructure:"worker"`
	RateLimit   *RateLimitConfig `mapstructure:"rate_limit"`
	Server      *ServerConfig    `mapstructure:"server"`
	Database    *DatabaseConfig  `mapstructure:"database"`
	Logger      *LoggerConfig    `mapstructure:"logger"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	Timeout         time.Duration `mapstructure:"timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Cors            *CORSConfig   `mapstructure:"cors"`
}

type CORSConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowedMethods   []string      `mapstructure:"allowed_methods"`
	AllowedHeaders   []string      `mapstructure:"allowed_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
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
	Level string `mapstructure:"level"`
}

type WorkerConfig struct {
	JobBufferCapacity int `mapstructure:"job_buffer_capacity"`
	JobWorkerCount    int `mapstructure:"job_worker_count"`
	TaskWorkerCount   int `mapstructure:"task_worker_count"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `mapstructure:"requests_per_second"`
	Burst             int `mapstructure:"burst"`
}

func (c *ServerConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		// Format durations
		slog.String("read_timeout", c.ReadTimeout.String()),
		slog.String("write_timeout", c.WriteTimeout.String()),
		slog.String("idle_timeout", c.IdleTimeout.String()),
		slog.String("timeout", c.Timeout.String()),
		slog.String("shutdown_timeout", c.ShutdownTimeout.String()),
		slog.Any("cors", c.Cors),
	)
}

func (c *DatabaseConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.String("db_name", c.DBName),
		slog.Int("max_open_conns", c.MaxOpenConns),
		slog.Int("max_idle_conns", c.MaxIdleConns),
		slog.String("conn_max_lifetime", c.ConnMaxLifetime.String()),
	)
}

func (c *RateLimitConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("requests_per_second", c.RequestsPerSecond),
		slog.Int("burst", c.Burst),
	)
}

func (c *CORSConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("enabled", c.Enabled),
		slog.String("max_age", c.MaxAge.String()),
		slog.Any("allowed_origins", c.AllowedOrigins),
	)
}

// GetDatabaseDSN returns a PostgreSQL DSN connection string
func (config *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)
}

// GetServerAddress returns the full server address
func (config *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
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

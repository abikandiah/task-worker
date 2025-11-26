package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host            string           `mapstructure:"host"`
	Port            int              `mapstructure:"port"`
	ReadTimeout     time.Duration    `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration    `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration    `mapstructure:"idle_timeout"`
	Timeout         time.Duration    `mapstructure:"timeout"`
	ShutdownTimeout time.Duration    `mapstructure:"shutdown_timeout"`
	Cors            *CORSConfig      `mapstructure:"cors"`
	RateLimit       *RateLimitConfig `mapstructure:"rate_limit"`
}

type CORSConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowedMethods   []string      `mapstructure:"allowed_methods"`
	AllowedHeaders   []string      `mapstructure:"allowed_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `mapstructure:"requests_per_second"`
	Burst             int `mapstructure:"burst"`
}

func SetConfigDefaults(v *viper.Viper) {
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
	// --- Rate Limit Configuration Defaults ---
	v.SetDefault("server.rate_limit.requests_per_second", 100)
	v.SetDefault("server.rate_limit.burst", 200)
}

func BindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	v.BindEnv("server.timeout", "SERVER_TIMEOUT")
	v.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")
	// Rate Limit Config
	v.BindEnv("server.rate_limit.requests_per_second", "RATE_LIMIT_REQUESTS_PER_SECOND")
	v.BindEnv("server.rate_limit.burst", "RATE_LIMIT_BURST")
}

func (config *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (config *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", config.Host),
		slog.Int("port", config.Port),
		slog.Any("cors", config.Cors),
		slog.Any("rate_limit", config.RateLimit),
	)
}

func (config *Config) Validate() error {
	if config.ReadTimeout <= 0 || config.WriteTimeout <= 0 ||
		config.IdleTimeout <= 0 || config.ShutdownTimeout <= 0 ||
		config.Timeout <= 0 {

		return fmt.Errorf("server timeouts must be positive")
	}
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Port)
	}

	if err := config.RateLimit.Validate(); err != nil {
		return err
	}

	return nil
}

func (config *RateLimitConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("requests_per_second", config.RequestsPerSecond),
		slog.Int("burst", config.Burst),
	)
}

func (config *RateLimitConfig) Validate() error {
	if config.Burst < 1 || config.RequestsPerSecond < 1 {
		return fmt.Errorf("rate limits must be greater than 0")
	}
	return nil
}

func (config *CORSConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("enabled", config.Enabled),
		slog.String("max_age", config.MaxAge.String()),
		slog.Any("allowed_origins", config.AllowedOrigins),
	)
}

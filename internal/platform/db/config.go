package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
}

func SetConfigDefaults(v *viper.Viper) {
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.db_name", "myapp")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("database.conn_max_idle_time", 5*time.Minute)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.auto_migrate", true)
}

func BindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.db_name", "DB_NAME")
	v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	v.BindEnv("database.ssl_mode", "DB_SSL_MODE")
}

// DSN builds the database connection string from config
func (c Config) DSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
		)
	case "sqlite3":
		return c.DBName // For SQLite, DBName is the file path
	default:
		return ""
	}
}

func (config *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", config.Host),
		slog.Int("port", config.Port),
		slog.String("db_name", config.DBName),
		slog.String("conn_max_lifetime", config.ConnMaxLifetime.String()),
	)
}

func (config *Config) Validate() error {
	if config.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if config.Host == "" && config.Driver != "sqlite3" {
		return fmt.Errorf("database host is required")
	}
	if config.Port <= 0 && config.Driver != "sqlite3" {
		return fmt.Errorf("database port must be positive")
	}
	if config.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if config.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive")
	}
	if config.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be positive")
	}

	return nil
}

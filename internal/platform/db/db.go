// internal/platform/db/db.go
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/util"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/pressly/goose/v3"
)

type DB struct {
	*sqlx.DB
	driver string
	cfg    *Config
}

func New(config *Config) (*DB, error) {
	// Build DSN from config
	dsn := config.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	if config.Driver == "sqlite3" {
		if err := util.MakeDirs(config.DBName); err != nil {
			return nil, fmt.Errorf("error creating sqlite3 parent dirs: %w", err)
		}
	}

	// Open connection
	db, err := sqlx.Open(config.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Driver-specific configuration
	switch config.Driver {
	case "sqlite3":
		slog.Info("connected to sqlite3 db", "name", config.DBName)
		// SQLite-specific settings
		// SQLite only supports one writer at a time
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0) // SQLite connections don't need to be recycled

		// Enable SQLite pragmas for better performance and reliability
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pragmas := []string{
			"PRAGMA foreign_keys = ON",    // Enforce foreign key constraints
			"PRAGMA journal_mode = WAL",   // Write-Ahead Logging for better concurrency
			"PRAGMA synchronous = NORMAL", // Balance between safety and speed
			"PRAGMA busy_timeout = 5000",  // Wait up to 5 seconds on locked database
			"PRAGMA cache_size = -64000",  // Use 64MB of cache (negative = KB)
		}

		for _, pragma := range pragmas {
			if _, err := db.ExecContext(ctx, pragma); err != nil {
				db.Close()
				return nil, fmt.Errorf("execute pragma: %w", err)
			}
		}

	case "postgres":
		// PostgreSQL-specific settings
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	default:
		db.Close()
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{DB: db, driver: config.Driver, cfg: config}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Driver() string {
	return db.driver
}

func (db *DB) StatusCheck(ctx context.Context) error {
	const q = `SELECT 1`
	var tmp int
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// WithinTransaction executes a function within a database transaction
func (db *DB) WithinTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (db *DB) RunMigrations(migrationsDir string) error {
	if err := goose.SetDialect(db.driver); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db.DB.DB, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	slog.Info("migrations completed successfully", "driver", db.driver)
	return nil
}

func (db *DB) MigrateDown(migrationsDir string) error {
	if err := goose.SetDialect(db.driver); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Down(db.DB.DB, migrationsDir); err != nil {
		return fmt.Errorf("rollback migration: %w", err)
	}

	slog.Info("migration rolled back successfully")
	return nil
}

// MigrateStatus shows the current migration status
func (db *DB) MigrateStatus(migrationsDir string) error {
	if err := goose.SetDialect(db.driver); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Status(db.DB.DB, migrationsDir); err != nil {
		return fmt.Errorf("get migration status: %w", err)
	}

	return nil
}

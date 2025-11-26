// internal/platform/db/db.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB wraps sqlx.DB with additional functionality
type DB struct {
	*sqlx.DB
	driver string
	cfg    Config
}

// New creates a new database connection with sqlx for enhanced features
func New(cfg Config) (*DB, error) {
	// Open connection
	db, err := sqlx.Open(cfg.Driver, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{DB: db, driver: cfg.Driver, cfg: cfg}, nil
}

// Close gracefully closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Driver returns the database driver name
func (db *DB) Driver() string {
	return db.driver
}

// StatusCheck verifies database connectivity with context
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

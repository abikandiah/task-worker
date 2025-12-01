package postgres

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PostgresServiceRepository struct {
	DB *sqlx.DB
}

func NewPostgresServiceRepository(db *sqlx.DB) *PostgresServiceRepository {
	return &PostgresServiceRepository{
		DB: db,
	}
}

func (repo *PostgresServiceRepository) Close() error {
	return repo.DB.Close()
}

func isPostgresUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL returns a specific error code for unique violations
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// Code 23505 is the PostgreSQL error code for unique_violation
		return pqErr.Code == "23505"
	}

	return false
}

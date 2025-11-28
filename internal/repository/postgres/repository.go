package postgres

import (
	"github.com/jmoiron/sqlx"
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

package sqlite3

import (
	"strings"

	"github.com/jmoiron/sqlx"
)

type SQLiteServiceRepository struct {
	DB *sqlx.DB
}

func NewSQLiteServiceRepository(db *sqlx.DB) *SQLiteServiceRepository {
	return &SQLiteServiceRepository{
		DB: db,
	}
}

func (repo *SQLiteServiceRepository) Close() error {
	return repo.DB.Close()
}

func isSQLiteUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite unique constraint errors often contain this specific phrase.
	// The table name and index/column name might also be present in the full message.
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

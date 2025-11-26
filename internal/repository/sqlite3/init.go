package sqlite3

import (
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

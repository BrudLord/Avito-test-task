package repository

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return Repository{db: db}
}

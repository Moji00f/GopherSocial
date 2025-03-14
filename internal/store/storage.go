package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostStore{db},
	}
}

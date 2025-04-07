package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
	ErrDuplicateUsername = errors.New("a user with that username already exists")
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}

	Comments interface {
		GetByPostId(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}

	User interface {
		Create(context.Context, *sql.Tx, *User) error
		GetById(context.Context, int64) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
	}

	Followers interface {
		Follow(ctx context.Context, followerId, userId int64) error
		UnFollow(ctx context.Context, followerId, userId int64) error
	}

	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		User:      &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowStore{db},
		Roles:     &RoleStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

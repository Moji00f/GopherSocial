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
		Create(context.Context, *User) error
		GetById(context.Context, int64) (*User, error)
	}

	Followers interface {
		Follow(ctx context.Context, followerId, userId int64) error
		UnFollow(ctx context.Context, followerId, userId int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		User:      &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowStore{db},
	}
}

package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type Follower struct {
	UserId     int64  `json:"user_id"`
	FollowerId int64  `json:"follow_id"`
	CreatedAt  string `json:"created_at"`
}
type FollowStore struct {
	db *sql.DB
}

func (f *FollowStore) Follow(ctx context.Context, followerId, userId int64) error {
	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1,$2)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := f.db.ExecContext(ctx, query, userId, followerId)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return nil
}

func (f *FollowStore) UnFollow(ctx context.Context, followerId, userId int64) error {
	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id=$2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := f.db.ExecContext(ctx, query, userId, followerId)

	return err
}

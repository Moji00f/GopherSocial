package store

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username, password, email) VALUES ($1,$2,$3) RETURNING id, created_at
	`

	err := u.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) GetById(ctx context.Context, postId int64) (*User, error) {
	query := `
			SELECT id, username, email, password, created_at FROM users WHERE id=$1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}

	err := u.db.QueryRowContext(ctx, query, postId).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

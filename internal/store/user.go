package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type UserStore struct {
	db *sql.DB
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

func (u *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, password, email) VALUES ($1,$2,$3) RETURNING id, created_at
	`

	err := tx.QueryRowContext(ctx, query, user.Username, user.Password.hash, user.Email).Scan(&user.ID, &user.CreatedAt)

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Constraint {
		//pq: duplicate key value violates unique constraint "users_username_key"
		case "users_username_key":
			return ErrDuplicateUsername
		//pq: duplicate key value violates unique constraint "users_email_key"
		case "users_email_key":
			return ErrDuplicateEmail
		}
	}

	return err
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

func (u *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	// transaction wrapper
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}

		// create invitation
		err := u.createUserAndInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}
		return nil
	})

}

func (u *UserStore) createUserAndInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userId int64) error {
	query := `INSERT INTO user_invitaions (token, user_id, expiry) VALUES ($1,$2,$3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userId, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil

}

func (u *UserStore) Activate(ctx context.Context, token string) error {

	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// 1. find the user that this token belongs to
		user, err := u.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		// 2. update the user
		user.IsActive = true
		if err := u.update(ctx, tx, user); err != nil {
			return err
		}
		// 3. clean the invitations
		if err := u.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})

}

func (u *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
				SELECT u.id, u.username, u.email, u.is_active, u.created_at
				FROM users u 
				JOIN user_invitaions ui ON u.id=ui.user_id 
				WHERE ui.token=$1 AND ui.expiry > $2
			`
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (u *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username=$1, email=$2, is_active=$3 WHERE id=$4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `DELETE FROM user_invitaions WHERE user_id =$1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) Delete(ctx context.Context, userId int64) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.delete(ctx, tx, userId); err != nil {
			return err
		}
		if err := u.deleteUserInvitations(ctx, tx, userId); err != nil {
			return err
		}

		return nil
	})
}
func (u *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id=$1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

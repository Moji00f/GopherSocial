package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentCounts int `json:"comment_counts"`
}

type PostStore struct {
	db *sql.DB
}

func (p *PostStore) Create(ctx context.Context, post *Post) error {
	query := ` 
      INSERT INTO posts (content, title, user_id, tags) VALUES ($1,$2,$3,$4)
 	  RETURNING id, created_at, updated_at;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *PostStore) GetById(ctx context.Context, id int64) (*Post, error) {

	query := `
		SELECT id, user_id, title, content, tags,version,created_at, updated_at FROM posts WHERE id=$1
	`

	var post Post

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (p *PostStore) Delete(ctx context.Context, postID int64) error {
	query := `DELETE FROM posts WHERE id=$1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := p.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts SET title=$1, content=$2, version=version+1 WHERE id = $3 AND version=$4 RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (p *PostStore) GetUserFeed(ctx context.Context, userId int64) ([]PostWithMetadata, error) {
	query := `
			SELECT 
	           p.id, p.user_id, p.title, p.content, p.created_at,p.version, p.tags,
               u.username, COUNT(c.id) AS comments_total
            FROM posts p 
            LEFT JOIN comments c ON c.post_id=p.id
            LEFT JOIN users u ON p.user_id = u.id
            LEFT JOIN followers f ON f.follower_id = p.user_id OR p.user_id=$1
            WHERE p.user_id=$1 OR f.follower_id=$1
            GROUP BY p.id, u.username
            ORDER BY p.created_at DESC;
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentCounts,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}

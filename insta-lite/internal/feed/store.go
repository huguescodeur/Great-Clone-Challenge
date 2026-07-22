package feed

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/post"
)

type FeedStore interface {
	GetFeed(ctx context.Context, followerID uuid.UUID, cursor *post.PostCursor, limit int) ([]*post.Post, error)
}

type store struct {
	db *sql.DB
}

func NewFeedStore(db *sql.DB) FeedStore {
	return &store{db: db}
}

func (s *store) GetFeed(ctx context.Context, followerID uuid.UUID, cursor *post.PostCursor, limit int) ([]*post.Post, error) {
	var q string
	var rows *sql.Rows
	var err error

	if cursor == nil {
		q = `
			SELECT p.post_id, p.user_id, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
			FROM posts p
			JOIN followers f ON f.followee_id = p.user_id
			WHERE f.follower_id = $1
			  AND f.status = 'accepted'
			  AND p.deleted_at IS NULL
			ORDER BY p.created_at DESC, p.post_id DESC
			LIMIT $2
			`
		rows, err = s.db.QueryContext(ctx, q, followerID, limit)
	} else {
		q = `
			SELECT p.post_id, p.user_id, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
			FROM posts p
			JOIN followers f ON f.followee_id = p.user_id
			WHERE f.follower_id = $1
			  AND f.status = 'accepted'
			  AND p.deleted_at IS NULL
			  AND (p.created_at < $2 OR (p.created_at = $3 AND p.post_id < $4))
			ORDER BY p.created_at DESC, p.post_id DESC
			LIMIT $5
			`
		rows, err = s.db.QueryContext(ctx, q, followerID, cursor.CreatedAt, cursor.CreatedAt, cursor.PostID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*post.Post
	for rows.Next() {
		p := post.Post{}
		if err := rows.Scan(
			&p.PostID,
			&p.UserID,
			&p.Content,
			&p.LikesCount,
			&p.CommentsCount,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

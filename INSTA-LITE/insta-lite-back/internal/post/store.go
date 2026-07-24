package post

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type PostStore interface {
	GetAll(ctx context.Context, cursor *PostCursor, limit int) ([]*Post, error)
	GetAllByID(ctx context.Context, userID uuid.UUID, cursor *PostCursor, limit int) ([]*Post, error)
	// GetByID(ctx context.Context, UserID, PostID uuid.UUID) (*Post, error)
	Update(ctx context.Context, content string, userID, postID uuid.UUID) (*Post, error)
	UpdateLikesCount(ctx context.Context, tx *sql.Tx, postID uuid.UUID, delta int) error
	UpdateCommentsCount(ctx context.Context, tx *sql.Tx, postID uuid.UUID, delta int) error
	Delete(ctx context.Context, tx *sql.Tx, userID, postID uuid.UUID) error
	Create(ctx context.Context, tx *sql.Tx, post *Post) (*Post, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Post, error)
	GetAuthorID(ctx context.Context, postID uuid.UUID) (uuid.UUID, error)
}

type store struct {
	db *sql.DB
}

func NewPostStore(db *sql.DB) PostStore {
	return &store{db: db}
}

func (s *store) GetAll(ctx context.Context, cursor *PostCursor, limit int) ([]*Post, error) {
	var q string
	var rows *sql.Rows
	var err error

	if cursor == nil {
		q = `
            SELECT p.post_id, p.user_id, u.username, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
            FROM posts p
            JOIN users u ON u.user_id = p.user_id
            WHERE p.deleted_at IS NULL
            ORDER BY p.created_at DESC, p.post_id DESC
            LIMIT $1`

		rows, err = s.db.QueryContext(ctx, q, limit)
	} else {
		q = `
            SELECT p.post_id, p.user_id, u.username, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
            FROM posts p
            JOIN users u ON u.user_id = p.user_id
            WHERE p.deleted_at IS NULL
              AND (p.created_at < $1 OR (p.created_at = $2 AND p.post_id < $3))
            ORDER BY p.created_at DESC, p.post_id DESC
            LIMIT $4`

		rows, err = s.db.QueryContext(ctx, q, cursor.CreatedAt, cursor.CreatedAt, cursor.PostID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := Post{}
		if err := rows.Scan(
			&p.PostID,
			&p.UserID,
			&p.Username,
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

func (s *store) GetAllByID(ctx context.Context, userID uuid.UUID, cursor *PostCursor, limit int) ([]*Post, error) {
	var q string
	var rows *sql.Rows
	var err error

	if cursor == nil {
		q = `
		SELECT p.post_id, p.user_id, u.username, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.user_id = p.user_id
		WHERE p.user_id = $1
		AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2
		`

		rows, err = s.db.QueryContext(ctx, q, userID, limit)
	} else {
		q = `
		SELECT p.post_id, p.user_id, u.username, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.user_id = p.user_id
		WHERE p.user_id = $1
		AND p.deleted_at IS NULL
		AND (p.created_at < $2 OR (p.created_at = $3 AND p.post_id < $4))
		ORDER BY p.created_at DESC
		LIMIT $5
		`

		rows, err = s.db.QueryContext(ctx, q, userID, cursor.CreatedAt, cursor.CreatedAt, cursor.PostID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		p := Post{}

		if err := rows.Scan(
			&p.PostID,
			&p.UserID,
			&p.Username,
			&p.Content,
			&p.LikesCount,
			&p.CommentsCount,
			&p.CreatedAt,
			&p.UpdatedAt); err != nil {
			return nil, err
		}

		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *store) Update(ctx context.Context, content string, userID, postID uuid.UUID) (*Post, error) {
	q := `
		UPDATE posts 
		SET content = $1, updated_at = NOW() 
		WHERE user_id = $2 
		AND post_id = $3 
		AND deleted_at IS NULL
		RETURNING post_id, user_id, content, likes_count, comments_count, created_at, updated_at
		`

	p := Post{}

	if err := s.db.QueryRowContext(ctx, q, content, userID, postID).Scan(
		&p.PostID,
		&p.UserID,
		&p.Content,
		&p.LikesCount,
		&p.CommentsCount,
		&p.CreatedAt,
		&p.UpdatedAt); err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *store) UpdateLikesCount(ctx context.Context, tx *sql.Tx, postID uuid.UUID, delta int) error {
	q := `
		UPDATE posts
		SET likes_count = GREATEST(likes_count + $1, 0)
		WHERE post_id = $2
		`

	if _, err := tx.ExecContext(ctx, q, delta, postID); err != nil {
		return err
	}

	return nil
}

func (s *store) UpdateCommentsCount(ctx context.Context, tx *sql.Tx, postID uuid.UUID, delta int) error {
	q := `
		UPDATE posts 
		SET comments_count = GREATEST(comments_count + $1, 0) 
		WHERE post_id = $2
		`

	_, err := tx.ExecContext(ctx, q, delta, postID)
	return err
}

func (s *store) Delete(ctx context.Context, tx *sql.Tx, userID, postID uuid.UUID) error {
	q := `
		UPDATE posts 
		SET deleted_at = NOW()
		WHERE user_id = $1 
		AND post_id = $2 
		AND deleted_at IS NULL
		`

	result, err := tx.ExecContext(ctx, q, userID, postID)
	if err != nil {
		return err
	}

	raw, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if raw == 0 {
		return fmt.Errorf("post n'existe pas ou déjà supprimé")
	}

	return nil
}

func (s *store) Create(ctx context.Context, tx *sql.Tx, post *Post) (*Post, error) {
	q := `
		INSERT INTO posts (post_id, user_id, content) 
		VALUES($1, $2, $3)
		RETURNING post_id, user_id, content, likes_count, comments_count, created_at, updated_at
		`

	if err := tx.QueryRowContext(ctx, q, post.PostID, post.UserID, post.Content).Scan(
		&post.PostID,
		&post.UserID,
		&post.Content,
		&post.LikesCount,
		&post.CommentsCount,
		&post.CreatedAt,
		&post.UpdatedAt); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *store) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Post, error) {
	if len(ids) == 0 {
		return []*Post{}, nil
	}

	q := `
		SELECT p.post_id, p.user_id, u.username, p.content, p.likes_count, p.comments_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.user_id = p.user_id
		WHERE p.post_id = ANY($1) AND p.deleted_at IS NULL
		`

	rows, err := s.db.QueryContext(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0, len(ids))
	for rows.Next() {
		p := Post{}
		if err := rows.Scan(
			&p.PostID, &p.UserID, &p.Username, &p.Content,
			&p.LikesCount, &p.CommentsCount,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}
	return posts, rows.Err()
}

func (s *store) GetAuthorID(ctx context.Context, postID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	q := `SELECT user_id FROM posts WHERE post_id = $1 AND deleted_at IS NULL`
	if err := s.db.QueryRowContext(ctx, q, postID).Scan(&userID); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

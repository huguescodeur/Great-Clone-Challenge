package comment

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type CommentStore interface {
	GetAllByPostID(ctx context.Context, postID uuid.UUID, cursor *CommentCursor, limit int) ([]*Comment, error)
	GetRepliesByParentID(ctx context.Context, parentCommentID uuid.UUID) ([]*Comment, error)
	GetByID(ctx context.Context, commentID uuid.UUID) (*Comment, error)
	Update(ctx context.Context, content string, userID, commentID uuid.UUID) (*Comment, error)
	Delete(ctx context.Context, tx *sql.Tx, userID, commentID uuid.UUID) error
	Create(ctx context.Context, tx *sql.Tx, comment *Comment) (*Comment, error)
	UpdateLikesCount(ctx context.Context, tx *sql.Tx, commentID uuid.UUID, delta int) error
	GetAuthorID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
}

type store struct {
	db *sql.DB
}

func NewCommentStore(db *sql.DB) CommentStore {
	return &store{db: db}
}

func (s *store) GetAllByPostID(ctx context.Context, postID uuid.UUID, cursor *CommentCursor, limit int) ([]*Comment, error) {
	var q string
	var rows *sql.Rows
	var err error

	if cursor == nil {
		q = `
			SELECT c.comment_id, c.post_id, c.user_id, u.username, c.parent_comment_id, c.content, c.likes_count, c.created_at, c.updated_at
			FROM comments c
			JOIN users u ON u.user_id = c.user_id
			WHERE c.post_id = $1 AND c.parent_comment_id IS NULL AND c.deleted_at IS NULL
			ORDER BY c.created_at DESC, c.comment_id DESC
			LIMIT $2
			`

		rows, err = s.db.QueryContext(ctx, q, postID, limit)
	} else {
		q = `
			SELECT c.comment_id, c.post_id, c.user_id, u.username, c.parent_comment_id, c.content, c.likes_count, c.created_at, c.updated_at
			FROM comments c
			JOIN users u ON u.user_id = c.user_id
			WHERE c.post_id = $1 AND c.parent_comment_id IS NULL AND c.deleted_at IS NULL
			AND (c.created_at < $2 OR (c.created_at = $3 AND c.comment_id < $4))
			ORDER BY c.created_at DESC, c.comment_id DESC
			LIMIT $5
			`

		rows, err = s.db.QueryContext(ctx, q, postID, cursor.CreatedAt, cursor.CreatedAt, cursor.CommentID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*Comment, 0)
	for rows.Next() {
		c := Comment{}
		if err := rows.Scan(
			&c.CommentID,
			&c.PostID,
			&c.UserID,
			&c.Username,
			&c.ParentCommentID,
			&c.Content,
			&c.LikesCount,
			&c.CreatedAt,
			&c.UpdatedAt); err != nil {
			return nil, err
		}

		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *store) GetRepliesByParentID(ctx context.Context, parentCommentID uuid.UUID) ([]*Comment, error) {
	q := `
		SELECT c.comment_id, c.post_id, c.user_id, u.username, c.parent_comment_id, c.content, c.likes_count, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON u.user_id = c.user_id
		WHERE c.parent_comment_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC
		`

	rows, err := s.db.QueryContext(ctx, q, parentCommentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := make([]*Comment, 0)
	for rows.Next() {
		c := Comment{}
		if err := rows.Scan(
			&c.CommentID,
			&c.PostID,
			&c.UserID,
			&c.Username,
			&c.ParentCommentID,
			&c.Content,
			&c.LikesCount,
			&c.CreatedAt,
			&c.UpdatedAt); err != nil {
			return nil, err
		}

		replies = append(replies, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return replies, nil
}

func (s *store) GetByID(ctx context.Context, commentID uuid.UUID) (*Comment, error) {
	q := `
		SELECT c.comment_id, c.post_id, c.user_id, u.username, c.parent_comment_id, c.content, c.likes_count, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON u.user_id = c.user_id
		WHERE c.comment_id = $1 AND c.deleted_at IS NULL
		`

	c := &Comment{}
	if err := s.db.QueryRowContext(ctx, q, commentID).Scan(
		&c.CommentID,
		&c.PostID,
		&c.UserID,
		&c.Username,
		&c.ParentCommentID,
		&c.Content,
		&c.LikesCount,
		&c.CreatedAt,
		&c.UpdatedAt); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *store) Update(ctx context.Context, content string, userID, commentID uuid.UUID) (*Comment, error) {
	q := `
		UPDATE comments
		SET content = $1, updated_at = NOW()
		WHERE comment_id = $2 AND user_id = $3 AND deleted_at IS NULL
		RETURNING comment_id, post_id, user_id, parent_comment_id, content, likes_count, created_at, updated_at
		`

	c := &Comment{}
	if err := s.db.QueryRowContext(ctx, q, content, commentID, userID).Scan(
		&c.CommentID,
		&c.PostID,
		&c.UserID,
		&c.ParentCommentID,
		&c.Content,
		&c.LikesCount,
		&c.CreatedAt,
		&c.UpdatedAt); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *store) Delete(ctx context.Context, tx *sql.Tx, userID, commentID uuid.UUID) error {
	q := `
		UPDATE comments
		SET deleted_at = NOW()
		WHERE comment_id = $1 AND user_id = $2 AND deleted_at IS NULL
		`

	result, err := tx.ExecContext(ctx, q, commentID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCommentNotFound
	}

	return nil
}

func (s *store) UpdateLikesCount(ctx context.Context, tx *sql.Tx, commentID uuid.UUID, delta int) error {
	q := `UPDATE comments SET likes_count = GREATEST(likes_count + $1, 0) WHERE comment_id = $2`

	if _, err := tx.ExecContext(ctx, q, delta, commentID); err != nil {
		return err
	}

	return nil
}

func (s *store) Create(ctx context.Context, tx *sql.Tx, comment *Comment) (*Comment, error) {
	q := `
		INSERT INTO comments (comment_id, post_id, user_id, parent_comment_id, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING comment_id, post_id, user_id, parent_comment_id, content, likes_count, created_at, updated_at
		`

	if err := tx.QueryRowContext(ctx, q,
		comment.CommentID, comment.PostID, comment.UserID, comment.ParentCommentID, comment.Content,
	).Scan(
		&comment.CommentID,
		&comment.PostID,
		&comment.UserID,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.LikesCount,
		&comment.CreatedAt,
		&comment.UpdatedAt); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *store) GetAuthorID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	q := `SELECT user_id FROM comments WHERE comment_id = $1 AND deleted_at IS NULL`
	if err := s.db.QueryRowContext(ctx, q, commentID).Scan(&userID); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

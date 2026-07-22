package commentlike

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/like"
)

type CommentLikeStore interface {
	GetAll(ctx context.Context, commentID uuid.UUID) ([]*CommentLike, error)
	GetByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*CommentLike, error)
	UpdateReaction(ctx context.Context, reactionType like.ReactionType, userID, commentID uuid.UUID) (*CommentLike, error)
	Delete(ctx context.Context, tx *sql.Tx, commentID, userID uuid.UUID) error
	Create(ctx context.Context, tx *sql.Tx, commentLike *CommentLike) (*CommentLike, error)
}

type store struct {
	db *sql.DB
}

func NewCommentLikeStore(db *sql.DB) CommentLikeStore {
	return &store{db: db}
}

func (s *store) GetAll(ctx context.Context, commentID uuid.UUID) ([]*CommentLike, error) {
	q := `
		SELECT comment_id, user_id, reaction_type, created_at, updated_at
		FROM comment_likes
		WHERE comment_id = $1
		ORDER BY created_at DESC
		`

	rows, err := s.db.QueryContext(ctx, q, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []*CommentLike
	for rows.Next() {
		l := CommentLike{}
		if err := rows.Scan(
			&l.CommentID,
			&l.UserID,
			&l.ReactionType,
			&l.Created_At,
			&l.Updated_At); err != nil {
			return nil, err
		}

		likes = append(likes, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return likes, nil
}

func (s *store) GetByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*CommentLike, error) {
	q := `
		SELECT comment_id, user_id, reaction_type, created_at, updated_at
		FROM comment_likes
		WHERE comment_id = $1 AND user_id = $2
		`

	l := &CommentLike{}

	if err := s.db.QueryRowContext(ctx, q, commentID, userID).Scan(
		&l.CommentID,
		&l.UserID,
		&l.ReactionType,
		&l.Created_At,
		&l.Updated_At); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *store) UpdateReaction(ctx context.Context, reactionType like.ReactionType, userID, commentID uuid.UUID) (*CommentLike, error) {
	q := `
		UPDATE comment_likes
		SET reaction_type = $1, updated_at = NOW()
		WHERE comment_id = $2 AND user_id = $3
		RETURNING comment_id, user_id, reaction_type, created_at, updated_at
		`

	l := &CommentLike{}

	if err := s.db.QueryRowContext(ctx, q, reactionType, commentID, userID).Scan(
		&l.CommentID,
		&l.UserID,
		&l.ReactionType,
		&l.Created_At,
		&l.Updated_At); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *store) Delete(ctx context.Context, tx *sql.Tx, commentID, userID uuid.UUID) error {
	q := `
		DELETE FROM comment_likes
		WHERE comment_id = $1 AND user_id = $2
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
		return ErrCommentLikeNotFound
	}

	return nil
}

func (s *store) Create(ctx context.Context, tx *sql.Tx, commentLike *CommentLike) (*CommentLike, error) {
	q := `
		INSERT INTO comment_likes (comment_id, user_id, reaction_type)
		VALUES ($1, $2, $3)
		RETURNING comment_id, user_id, reaction_type, created_at, updated_at
		`

	if err := tx.QueryRowContext(ctx, q, commentLike.CommentID, commentLike.UserID, commentLike.ReactionType).Scan(
		&commentLike.CommentID,
		&commentLike.UserID,
		&commentLike.ReactionType,
		&commentLike.Created_At,
		&commentLike.Updated_At); err != nil {
		return nil, err
	}

	return commentLike, nil
}

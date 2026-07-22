package like

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type LikeStore interface {
	// TODO: Pagination can be added later if needed
	GetAll(ctx context.Context, PostID uuid.UUID) ([]*Like, error)
	GetByUserAndPost(ctx context.Context, userId, PostID uuid.UUID) (*Like, error)
	UpdateReaction(ctx context.Context, reactionType ReactionType, userId, PostID uuid.UUID) (*Like, error)
	Delete(ctx context.Context, tx *sql.Tx, postID, userID uuid.UUID) error
	Create(ctx context.Context, tx *sql.Tx, like *Like) (*Like, error)
}

type store struct {
	db *sql.DB
}

func NewLikeStore(db *sql.DB) LikeStore {
	return &store{db: db}
}

func (s *store) GetAll(ctx context.Context, PostID uuid.UUID) ([]*Like, error) {
	q := `
		SELECT post_id, user_id, reaction_type, created_at, updated_at
		FROM likes
		WHERE post_id = $1
		ORDER BY created_at DESC
		`

	rows, err := s.db.QueryContext(ctx, q, PostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []*Like
	for rows.Next() {
		l := Like{}
		if err := rows.Scan(
			&l.PostID,
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

func (s *store) GetByUserAndPost(ctx context.Context, userId, postID uuid.UUID) (*Like, error) {
	q := `
		SELECT post_id, user_id, reaction_type, created_at, updated_at
		FROM likes
		WHERE post_id = $1 AND user_id = $2
		`

	l := &Like{}

	if err := s.db.QueryRowContext(ctx, q, postID, userId).Scan(
		&l.PostID,
		&l.UserID,
		&l.ReactionType,
		&l.Created_At,
		&l.Updated_At); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *store) UpdateReaction(ctx context.Context, reactionType ReactionType, userId, postID uuid.UUID) (*Like, error) {
	q := `
		UPDATE likes
		SET reaction_type = $1, updated_at = NOW()
		WHERE post_id = $2 AND user_id = $3 
		RETURNING post_id, user_id, reaction_type, created_at, updated_at
		`

	l := &Like{}

	if err := s.db.QueryRowContext(ctx, q, reactionType, postID, userId).Scan(
		&l.PostID,
		&l.UserID,
		&l.ReactionType,
		&l.Created_At,
		&l.Updated_At); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *store) Delete(ctx context.Context, tx *sql.Tx, postID, userID uuid.UUID) error {
	q := `
		DELETE FROM likes
		WHERE post_id = $1 AND user_id = $2 
		`

	result, err := tx.ExecContext(ctx, q, postID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrLikeNotFound
	}

	return nil
}

func (s *store) Create(ctx context.Context, tx *sql.Tx, like *Like) (*Like, error) {
	q := `
		INSERT INTO likes (post_id, user_id, reaction_type)
		VALUES ($1, $2, $3)
		RETURNING post_id, user_id, reaction_type, created_at, updated_at
		`

	if err := tx.QueryRowContext(ctx, q, like.PostID, like.UserID, like.ReactionType).Scan(
		&like.PostID,
		&like.UserID,
		&like.ReactionType,
		&like.Created_At,
		&like.Updated_At); err != nil {
		return nil, err
	}

	return like, nil
}

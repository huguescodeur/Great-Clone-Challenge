package user

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type UserStore interface {
	UpdatePostCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error)
	UpdateFollowersCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error)
	UpdateFollowingCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error)
}

type store struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) UserStore {
	return &store{db: db}
}

func (s *store) UpdatePostCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error) {
	q := `
		UPDATE users
		SET posts_count = GREATEST(posts_count + $1, 0)
		WHERE user_id = $2
		RETURNING posts_count
		`

	u := User{}

	if err := tx.QueryRowContext(ctx, q, delta, userID).Scan(&u.PostsCount); err != nil {
		return 0, err
	}

	return u.PostsCount, nil
}

func (s *store) UpdateFollowersCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error) {
	q := `
		UPDATE users
		SET followers_count = GREATEST(followers_count + $1, 0)
		WHERE user_id = $2
		RETURNING followers_count
		`

	u := User{}

	if err := tx.QueryRowContext(ctx, q, delta, userID).Scan(&u.FollowersCount); err != nil {
		return 0, err
	}

	return u.FollowersCount, nil
}

func (s *store) UpdateFollowingCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error) {
	q := `
		UPDATE users
		SET following_count = GREATEST(following_count + $1, 0)
		WHERE user_id = $2
		RETURNING following_count
		`

	u := User{}

	if err := tx.QueryRowContext(ctx, q, delta, userID).Scan(&u.FollowingCount); err != nil {
		return 0, err
	}

	return u.FollowingCount, nil
}

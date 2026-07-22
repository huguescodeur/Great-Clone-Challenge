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
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetFollowersCounts(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]int64, error)
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

func (s *store) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	q := `
		SELECT user_id, username, email, full_name, bio, profile_pic_url,
		       followers_count, following_count, posts_count, created_at
		FROM users
		WHERE user_id = $1
		`

	u := User{}
	if err := s.db.QueryRowContext(ctx, q, userID).Scan(
		&u.UserID, &u.Username, &u.Email, &u.FullName, &u.Bio, &u.ProfilePicURL,
		&u.FollowersCount, &u.FollowingCount, &u.PostsCount, &u.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *store) GetFollowersCounts(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]int64{}, nil
	}

	q := `SELECT user_id, followers_count FROM users WHERE user_id = ANY($1)`

	rows, err := s.db.QueryContext(ctx, q, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[uuid.UUID]int64, len(userIDs))
	for rows.Next() {
		var id uuid.UUID
		var count int64
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		counts[id] = count
	}
	return counts, rows.Err()
}

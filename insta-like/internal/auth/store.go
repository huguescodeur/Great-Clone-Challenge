package auth

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-like/internal/user"
)

type AuthStore interface {
	Create(ctx context.Context, user *user.User) (*user.User, error)
	GetByEmailOrUsername(ctx context.Context, identifier string) (*user.User, error)
	UpdatePassword(ctx context.Context, newHash string, id uuid.UUID) error
}

type store struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) AuthStore {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, user *user.User) (*user.User, error) {
	q := `
	INSERT INTO users (
		user_id,
		username, 
		email, 
		password_hash, 
		full_name, 
		bio, 
		profile_pic_url
		)
	VALUES($1, $2, $3, $4, $5, $6, $7)
	RETURNING followers_count, following_count, posts_count, created_at
		`

	// RETURNING user_id, created_at, updated_at

	if err := s.db.QueryRowContext(
		ctx, q,
		user.UserID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		"",
		"",
	).Scan(&user.FollowersCount, &user.FollowingCount, &user.PostsCount, &user.CreatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *store) GetByEmailOrUsername(ctx context.Context, identifier string) (*user.User, error) {
	u := user.User{}

	q := `
		SELECT 
		user_id,
		username,
		email,
		password_hash,
		full_name,
		bio,
		profile_pic_url,
		followers_count,
		following_count,
		posts_count,
		created_at
		FROM users
		WHERE email = $1 OR username = $1
		LIMIT 1;
		`

	if err := s.db.QueryRowContext(ctx, q, identifier).Scan(
		&u.UserID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Bio,
		&u.ProfilePicURL,
		&u.FollowersCount,
		&u.FollowingCount,
		&u.PostsCount,
		&u.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *store) UpdatePassword(ctx context.Context, newHash string, id uuid.UUID) error {
	q := `UPDATE users SET password_hash = $1 WHERE user_id = $2`

	raw, err := s.db.ExecContext(ctx, q, newHash, id)
	if err != nil {
		return err
	}

	val, err := raw.RowsAffected()
	if err != nil {
		return err
	}

	if val == 0 {
		return ErrUserNotFound
	}

	return nil
}

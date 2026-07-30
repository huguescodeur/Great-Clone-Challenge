package follow

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type FollowStore interface {
	Create(ctx context.Context, tx *sql.Tx, f *Follow) (*Follow, error)
	Delete(ctx context.Context, tx *sql.Tx, followerID, followeeID uuid.UUID) error
	GetStatus(ctx context.Context, followerID, followeeID uuid.UUID) (*Follow, error)
	GetFollowers(ctx context.Context, followeeID uuid.UUID, limit, offset int) ([]*Follow, int, error)
	GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*Follow, int, error)
	GetFollowerIDs(ctx context.Context, followeeID uuid.UUID) ([]uuid.UUID, error)
	GetFolloweeIDs(ctx context.Context, followerID uuid.UUID) ([]uuid.UUID, error)
}

type store struct {
	db *sql.DB
}

func NewFollowStore(db *sql.DB) FollowStore {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, tx *sql.Tx, f *Follow) (*Follow, error) {
	q := `
		INSERT INTO followers (follower_id, followee_id, status)
		VALUES ($1, $2, $3)
		RETURNING follower_id, followee_id, status, created_at
		`

	if err := tx.QueryRowContext(ctx, q, f.FollowerID, f.FolloweeID, f.Status).Scan(
		&f.FollowerID,
		&f.FolloweeID,
		&f.Status,
		&f.CreatedAt); err != nil {
		return nil, err
	}

	return f, nil
}

func (s *store) Delete(ctx context.Context, tx *sql.Tx, followerID, followeeID uuid.UUID) error {
	q := `
		DELETE FROM followers
		WHERE follower_id = $1 AND followee_id = $2
		`

	result, err := tx.ExecContext(ctx, q, followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}

func (s *store) GetStatus(ctx context.Context, followerID, followeeID uuid.UUID) (*Follow, error) {
	q := `
		SELECT follower_id, followee_id, status, created_at
		FROM followers
		WHERE follower_id = $1 AND followee_id = $2
		`

	f := &Follow{}
	if err := s.db.QueryRowContext(ctx, q, followerID, followeeID).Scan(
		&f.FollowerID,
		&f.FolloweeID,
		&f.Status,
		&f.CreatedAt); err != nil {
		return nil, err
	}

	return f, nil
}

func (s *store) GetFollowers(ctx context.Context, followeeID uuid.UUID, limit, offset int) ([]*Follow, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM followers WHERE followee_id = $1 AND status = 'accepted'`,
		followeeID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `
		SELECT f.follower_id, f.followee_id, u.username, f.status, f.created_at
		FROM followers f
		JOIN users u ON u.user_id = f.follower_id
		WHERE f.followee_id = $1 AND f.status = 'accepted'
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
		`

	rows, err := s.db.QueryContext(ctx, q, followeeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	followers := make([]*Follow, 0)
	for rows.Next() {
		f := Follow{}
		if err := rows.Scan(&f.FollowerID, &f.FolloweeID, &f.FollowerUsername, &f.Status, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		followers = append(followers, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return followers, total, nil
}

func (s *store) GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*Follow, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM followers WHERE follower_id = $1 AND status = 'accepted'`,
		followerID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `
		SELECT f.follower_id, f.followee_id, u.username, f.status, f.created_at
		FROM followers f
		JOIN users u ON u.user_id = f.followee_id
		WHERE f.follower_id = $1 AND f.status = 'accepted'
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
		`

	rows, err := s.db.QueryContext(ctx, q, followerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	following := make([]*Follow, 0)
	for rows.Next() {
		f := Follow{}
		if err := rows.Scan(&f.FollowerID, &f.FolloweeID, &f.FolloweeUsername, &f.Status, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		following = append(following, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return following, total, nil
}

func (s *store) GetFollowerIDs(ctx context.Context, followeeID uuid.UUID) ([]uuid.UUID, error) {
	q := `
		SELECT follower_id
		FROM followers
		WHERE followee_id = $1 AND status = 'accepted'
		`

	rows, err := s.db.QueryContext(ctx, q, followeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *store) GetFolloweeIDs(ctx context.Context, followerID uuid.UUID) ([]uuid.UUID, error) {
	q := `
		SELECT followee_id
		FROM followers
		WHERE follower_id = $1 AND status = 'accepted'
		`

	rows, err := s.db.QueryContext(ctx, q, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

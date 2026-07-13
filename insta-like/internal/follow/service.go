package follow

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type FollowUpdater interface {
	UpdateFollowersCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error)
	UpdateFollowingCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta int) (int64, error)
}

type FollowService struct {
	followStore   FollowStore
	followUpdater FollowUpdater
	db            *sql.DB
}

func NewServiceFollow(followStore FollowStore, followUpdater FollowUpdater, db *sql.DB) *FollowService {
	return &FollowService{followStore: followStore, followUpdater: followUpdater, db: db}
}

func (s *FollowService) CreateFollow(ctx context.Context, followerID, followeeID uuid.UUID) (*Follow, error) {
	if followerID == followeeID {
		return nil, ErrCannotFollowSelf
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	f := &Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		Status:     StatusAccepted,
	}

	created, err := s.followStore.Create(ctx, tx, f)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyFollowing
		}
		return nil, ErrCreateFollowFailed
	}

	if _, err := s.followUpdater.UpdateFollowersCount(ctx, tx, followeeID, 1); err != nil {
		return nil, err
	}
	if _, err := s.followUpdater.UpdateFollowingCount(ctx, tx, followerID, 1); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.followStore.Delete(ctx, tx, followerID, followeeID); err != nil {
		return err
	}

	if _, err := s.followUpdater.UpdateFollowersCount(ctx, tx, followeeID, -1); err != nil {
		return err
	}
	if _, err := s.followUpdater.UpdateFollowingCount(ctx, tx, followerID, -1); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FollowService) GetStatus(ctx context.Context, followerID, followeeID uuid.UUID) (*Follow, error) {
	f, err := s.followStore.GetStatus(ctx, followerID, followeeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFollowNotFound
		}
		return nil, err
	}

	return f, nil
}

func (s *FollowService) GetFollowers(ctx context.Context, followeeID uuid.UUID, limit, offset int) ([]*Follow, int, error) {
	return s.followStore.GetFollowers(ctx, followeeID, limit, offset)
}

func (s *FollowService) GetFollowing(ctx context.Context, followerID uuid.UUID, limit, offset int) ([]*Follow, int, error) {
	return s.followStore.GetFollowing(ctx, followerID, limit, offset)
}

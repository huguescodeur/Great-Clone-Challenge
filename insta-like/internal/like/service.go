package like

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostUpdater interface {
	UpdateLikesCount(ctx context.Context, tx *sql.Tx, postID uuid.UUID, delta int) error
}

type LikeService struct {
	likeStore   LikeStore
	postUpdater PostUpdater
	// postStore post.PostStore
	db *sql.DB
}

func NewServiceLike(likeStore LikeStore, postUpdater PostUpdater, db *sql.DB) *LikeService {
	return &LikeService{likeStore: likeStore, postUpdater: postUpdater, db: db}
}

func (s *LikeService) GetAllLikes(ctx context.Context, postID uuid.UUID) ([]*Like, error) {
	likes, err := s.likeStore.GetAll(ctx, postID)
	if err != nil {
		return nil, err
	}

	return likes, nil
}

func (s *LikeService) GetLikeByUserAndPost(ctx context.Context, userId, postID uuid.UUID) (*Like, error) {
	likes, err := s.likeStore.GetByUserAndPost(ctx, userId, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReactionNotFound
		}

		return nil, err
	}

	return likes, nil
}

func (s *LikeService) UpdateReaction(ctx context.Context, reactionType ReactionType, userId, postID uuid.UUID) (*Like, error) {
	if !ReactionType.IsValid(reactionType) {
		return nil, ErrInvalidReactionType
	}

	like, err := s.likeStore.UpdateReaction(ctx, reactionType, userId, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReactionNotFound
		}

		return nil, err
	}

	return like, nil
}

func (s *LikeService) Delete(ctx context.Context, postID, userID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.likeStore.Delete(ctx, tx, postID, userID); err != nil {
		if errors.Is(err, ErrLikeNotFound) {
			return ErrReactionNotFound
		}

		return err
	}

	if err := s.postUpdater.UpdateLikesCount(ctx, tx, postID, -1); err != nil {
		return ErrUpdateCountFailed
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *LikeService) CreateLike(ctx context.Context, like Like) (*Like, error) {
	if !ReactionType.IsValid(like.ReactionType) {
		return nil, ErrInvalidReactionType
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	likeCreated, err := s.likeStore.Create(ctx, tx, &like)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyLiked
		}

		return nil, ErrCreateReactionFailed
	}

	if err := s.postUpdater.UpdateLikesCount(ctx, tx, like.PostID, 1); err != nil {

		return nil, ErrUpdateCountFailed
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return likeCreated, nil
}

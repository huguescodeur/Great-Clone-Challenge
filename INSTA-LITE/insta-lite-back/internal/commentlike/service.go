package commentlike

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/events"
	"github.com/huguescodeur/insta-lite/internal/like"
	"github.com/jackc/pgx/v5/pgconn"
)

type CommentUpdater interface {
	UpdateLikesCount(ctx context.Context, tx *sql.Tx, commentID uuid.UUID, delta int) error
}

type CommentAuthorGetter interface {
	GetAuthorID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
}

type CommentLikeService struct {
	commentLikeStore CommentLikeStore
	commentUpdater   CommentUpdater
	counter          CommentLikesCounter
	commentAuthor    CommentAuthorGetter
	bus              *events.TypedBus[events.CommentLikedEvent]
	db               *sql.DB
}

func NewServiceCommentLike(commentLikeStore CommentLikeStore, commentUpdater CommentUpdater, counter CommentLikesCounter, commentAuthor CommentAuthorGetter, bus *events.TypedBus[events.CommentLikedEvent], db *sql.DB) *CommentLikeService {
	return &CommentLikeService{commentLikeStore: commentLikeStore, commentUpdater: commentUpdater, counter: counter, commentAuthor: commentAuthor, bus: bus, db: db}
}

func (s *CommentLikeService) GetAllLikes(ctx context.Context, commentID uuid.UUID) ([]*CommentLike, error) {
	return s.commentLikeStore.GetAll(ctx, commentID)
}

func (s *CommentLikeService) GetLikeByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*CommentLike, error) {
	l, err := s.commentLikeStore.GetByUserAndComment(ctx, userID, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReactionNotFound
		}
		return nil, err
	}

	return l, nil
}

func (s *CommentLikeService) UpdateReaction(ctx context.Context, reactionType like.ReactionType, userID, commentID uuid.UUID) (*CommentLike, error) {
	if !reactionType.IsValid() {
		return nil, ErrInvalidReactionType
	}

	l, err := s.commentLikeStore.UpdateReaction(ctx, reactionType, userID, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReactionNotFound
		}
		return nil, err
	}

	return l, nil
}

func (s *CommentLikeService) Delete(ctx context.Context, commentID, userID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.commentLikeStore.Delete(ctx, tx, commentID, userID); err != nil {
		if errors.Is(err, ErrCommentLikeNotFound) {
			return ErrReactionNotFound
		}
		return err
	}

	// if err := s.commentUpdater.UpdateLikesCount(ctx, tx, commentID, -1); err != nil {
	// 	return ErrUpdateCountFailed
	// }

	// return tx.Commit()
	if err := tx.Commit(); err != nil {
		return err
	}

	if _, err := s.counter.Decrement(ctx, commentID); err != nil {
		log.Printf("[commentlike] échec décrément Redis commentID=%s: %v", commentID, err)
	}
	return nil

}

func (s *CommentLikeService) CreateLike(ctx context.Context, commentLike CommentLike) (*CommentLike, error) {
	if !commentLike.ReactionType.IsValid() {
		return nil, ErrInvalidReactionType
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	created, err := s.commentLikeStore.Create(ctx, tx, &commentLike)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyLiked
		}
		return nil, ErrCreateReactionFailed
	}

	// if err := s.commentUpdater.UpdateLikesCount(ctx, tx, commentLike.CommentID, 1); err != nil {
	// 	return nil, ErrUpdateCountFailed
	// }

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if _, err := s.counter.Increment(ctx, commentLike.CommentID); err != nil {
		log.Printf("[commentlike] échec incrément Redis commentID=%s: %v", commentLike.CommentID, err)
	}

	authorID, err := s.commentAuthor.GetAuthorID(ctx, commentLike.CommentID)
	if err != nil {
		log.Printf("[commentlike] échec récupération auteur commentID=%s: %v", commentLike.CommentID, err)
		return created, nil
	}

	if authorID != commentLike.UserID {
		s.bus.Publish(events.CommentLikedEvent{
			CommentID:       commentLike.CommentID,
			CommentAuthorID: authorID,
			ActorID:         commentLike.UserID,
			CreatedAt:       created.Created_At,
		})
	}

	return created, nil
}

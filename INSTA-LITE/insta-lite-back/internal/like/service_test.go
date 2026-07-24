package like_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/events"
	"github.com/huguescodeur/insta-lite/internal/like"
)

// --- stubs ---

type stubLikeStore struct {
	createFn func(ctx context.Context, tx *sql.Tx, l *like.Like) (*like.Like, error)
	deleteFn func(ctx context.Context, tx *sql.Tx, postID, userID uuid.UUID) error
}

func (s *stubLikeStore) Create(ctx context.Context, tx *sql.Tx, l *like.Like) (*like.Like, error) {
	return s.createFn(ctx, tx, l)
}
func (s *stubLikeStore) Delete(ctx context.Context, tx *sql.Tx, postID, userID uuid.UUID) error {
	return s.deleteFn(ctx, tx, postID, userID)
}
func (s *stubLikeStore) GetAll(_ context.Context, _ uuid.UUID) ([]*like.Like, error) {
	return nil, nil
}
func (s *stubLikeStore) GetByUserAndPost(_ context.Context, _, _ uuid.UUID) (*like.Like, error) {
	return nil, nil
}
func (s *stubLikeStore) UpdateReaction(_ context.Context, _ like.ReactionType, _, _ uuid.UUID) (*like.Like, error) {
	return nil, nil
}

type stubPostUpdater struct{}

func (s *stubPostUpdater) UpdateLikesCount(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ int) error {
	return nil
}

type stubLikeCounter struct {
	incrementFn func(ctx context.Context, postID uuid.UUID) (int64, error)
	decrementFn func(ctx context.Context, postID uuid.UUID) (int64, error)
}

func (s *stubLikeCounter) Increment(ctx context.Context, postID uuid.UUID) (int64, error) {
	if s.incrementFn != nil {
		return s.incrementFn(ctx, postID)
	}
	return 1, nil
}
func (s *stubLikeCounter) Decrement(ctx context.Context, postID uuid.UUID) (int64, error) {
	if s.decrementFn != nil {
		return s.decrementFn(ctx, postID)
	}
	return 0, nil
}
func (s *stubLikeCounter) GetCounts(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]int64, error) {
	return nil, nil
}

type stubAuthorGetter struct {
	getAuthorIDFn func(ctx context.Context, postID uuid.UUID) (uuid.UUID, error)
}

func (s *stubAuthorGetter) GetAuthorID(ctx context.Context, postID uuid.UUID) (uuid.UUID, error) {
	return s.getAuthorIDFn(ctx, postID)
}

// --- helpers ---

func newBus() *events.TypedBus[events.PostLikedEvent] {
	return events.NewTypedBus[events.PostLikedEvent](10)
}

// --- tests ---

func TestCreateLike_InvalidReactionType_ReturnsError(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	svc := like.NewServiceLike(
		&stubLikeStore{},
		&stubPostUpdater{},
		&stubLikeCounter{},
		&stubAuthorGetter{},
		newBus(),
		db,
	)

	_, err := svc.CreateLike(context.Background(), like.Like{
		PostID:       uuid.New(),
		UserID:       uuid.New(),
		ReactionType: "invalid",
	})

	if !errors.Is(err, like.ErrInvalidReactionType) {
		t.Errorf("want ErrInvalidReactionType, got %v", err)
	}
}

func TestCreateLike_Success_PublishesEventWhenNotSelfLike(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()
	authorID := uuid.New() // different from userID

	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	likeStore := &stubLikeStore{
		createFn: func(_ context.Context, _ *sql.Tx, l *like.Like) (*like.Like, error) {
			return &like.Like{
				PostID:       l.PostID,
				UserID:       l.UserID,
				ReactionType: l.ReactionType,
				Created_At:   time.Now(),
			}, nil
		},
	}
	authorGetter := &stubAuthorGetter{
		getAuthorIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return authorID, nil
		},
	}
	bus := newBus()

	svc := like.NewServiceLike(likeStore, &stubPostUpdater{}, &stubLikeCounter{}, authorGetter, bus, db)

	result, err := svc.CreateLike(context.Background(), like.Like{
		PostID:       postID,
		UserID:       userID,
		ReactionType: like.ReactionLike,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// event should be on the bus
	select {
	case e := <-bus.Events():
		if e.PostID != postID {
			t.Errorf("want postID %s, got %s", postID, e.PostID)
		}
		if e.ActorID != userID {
			t.Errorf("want actorID %s, got %s", userID, e.ActorID)
		}
		if e.PostAuthorID != authorID {
			t.Errorf("want authorID %s, got %s", authorID, e.PostAuthorID)
		}
	default:
		t.Error("expected event on bus, but bus was empty")
	}
}

func TestCreateLike_SelfLike_NoEventPublished(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	likeStore := &stubLikeStore{
		createFn: func(_ context.Context, _ *sql.Tx, l *like.Like) (*like.Like, error) {
			return &like.Like{PostID: l.PostID, UserID: l.UserID, ReactionType: l.ReactionType}, nil
		},
	}
	authorGetter := &stubAuthorGetter{
		getAuthorIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil // author == actor
		},
	}
	bus := newBus()

	svc := like.NewServiceLike(likeStore, &stubPostUpdater{}, &stubLikeCounter{}, authorGetter, bus, db)
	_, err := svc.CreateLike(context.Background(), like.Like{
		PostID:       postID,
		UserID:       userID,
		ReactionType: like.ReactionLike,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-bus.Events():
		t.Error("expected no event for self-like, but got one")
	default:
		// correct: bus is empty
	}
}

func TestDelete_Success_DecrementsCounter(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	decremented := false
	likeStore := &stubLikeStore{
		deleteFn: func(_ context.Context, _ *sql.Tx, _, _ uuid.UUID) error { return nil },
	}
	counter := &stubLikeCounter{
		decrementFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			decremented = true
			return 0, nil
		},
	}

	svc := like.NewServiceLike(likeStore, &stubPostUpdater{}, counter, &stubAuthorGetter{}, newBus(), db)
	err := svc.Delete(context.Background(), postID, userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decremented {
		t.Error("expected counter.Decrement to be called")
	}
}

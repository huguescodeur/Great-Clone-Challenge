package feed_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/events"
	"github.com/huguescodeur/insta-lite/internal/feed"
	"github.com/huguescodeur/insta-lite/internal/follow"
	"github.com/huguescodeur/insta-lite/internal/post"
	"github.com/huguescodeur/insta-lite/internal/user"
)

// --- stubs ---

type stubUserStore struct {
	getByIDFn          func(ctx context.Context, userID uuid.UUID) (*user.User, error)
	getFollowersCountsFn func(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

func (s *stubUserStore) GetByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	return s.getByIDFn(ctx, userID)
}
func (s *stubUserStore) UpdatePostCount(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ int) (int64, error) {
	return 0, nil
}
func (s *stubUserStore) UpdateFollowersCount(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ int) (int64, error) {
	return 0, nil
}
func (s *stubUserStore) UpdateFollowingCount(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ int) (int64, error) {
	return 0, nil
}
func (s *stubUserStore) GetFollowersCounts(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if s.getFollowersCountsFn != nil {
		return s.getFollowersCountsFn(ctx, userIDs)
	}
	return nil, nil
}

type stubFollowStore struct {
	getFollowerIDsFn func(ctx context.Context, followeeID uuid.UUID) ([]uuid.UUID, error)
}

func (s *stubFollowStore) GetFollowerIDs(ctx context.Context, followeeID uuid.UUID) ([]uuid.UUID, error) {
	return s.getFollowerIDsFn(ctx, followeeID)
}
func (s *stubFollowStore) Create(_ context.Context, _ *sql.Tx, _ *follow.Follow) (*follow.Follow, error) {
	return nil, nil
}
func (s *stubFollowStore) Delete(_ context.Context, _ *sql.Tx, _, _ uuid.UUID) error { return nil }
func (s *stubFollowStore) GetStatus(_ context.Context, _, _ uuid.UUID) (*follow.Follow, error) {
	return nil, nil
}
func (s *stubFollowStore) GetFollowers(_ context.Context, _ uuid.UUID, _, _ int) ([]*follow.Follow, int, error) {
	return nil, 0, nil
}
func (s *stubFollowStore) GetFollowing(_ context.Context, _ uuid.UUID, _, _ int) ([]*follow.Follow, int, error) {
	return nil, 0, nil
}
func (s *stubFollowStore) GetFolloweeIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type stubFeedCache struct {
	addPostFn              func(ctx context.Context, userID, postID uuid.UUID, t time.Time) error
	addToCelebrityFeedFn   func(ctx context.Context, authorID, postID uuid.UUID, t time.Time) error
	addPostCalled          []uuid.UUID // follower IDs that received the push
	celebrityFeedCalled    bool
}

func (s *stubFeedCache) AddPost(ctx context.Context, userID, postID uuid.UUID, t time.Time) error {
	s.addPostCalled = append(s.addPostCalled, userID)
	if s.addPostFn != nil {
		return s.addPostFn(ctx, userID, postID, t)
	}
	return nil
}
func (s *stubFeedCache) AddToCelebrityFeed(ctx context.Context, authorID, postID uuid.UUID, t time.Time) error {
	s.celebrityFeedCalled = true
	if s.addToCelebrityFeedFn != nil {
		return s.addToCelebrityFeedFn(ctx, authorID, postID, t)
	}
	return nil
}
func (s *stubFeedCache) GetFeed(_ context.Context, _ uuid.UUID, _ *post.PostCursor, _ int) ([]feed.FeedEntry, error) {
	return nil, nil
}
func (s *stubFeedCache) WarmCache(_ context.Context, _ uuid.UUID, _ []*post.Post) error { return nil }
func (s *stubFeedCache) GetCelebrityFeed(_ context.Context, _ uuid.UUID, _ *post.PostCursor, _ int) ([]feed.FeedEntry, error) {
	return nil, nil
}

// --- tests ---

func TestFanout_PushMode_DistributesToAllFollowers(t *testing.T) {
	authorID := uuid.New()
	follower1, follower2 := uuid.New(), uuid.New()

	userStore := &stubUserStore{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*user.User, error) {
			return &user.User{FollowersCount: 1}, nil // below pushThreshold (2)
		},
	}
	followStore := &stubFollowStore{
		getFollowerIDsFn: func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{follower1, follower2}, nil
		},
	}
	cache := &stubFeedCache{}

	f := feed.NewFanout(followStore, userStore, cache)
	f.HandlePostCreated(context.Background(), events.PostCreatedEvent{
		PostID:    uuid.New(),
		UserID:    authorID,
		CreatedAt: time.Now(),
	})

	if len(cache.addPostCalled) != 2 {
		t.Errorf("expected 2 AddPost calls (push mode), got %d", len(cache.addPostCalled))
	}
	if cache.celebrityFeedCalled {
		t.Error("expected celebrity feed NOT to be called in push mode")
	}
}

func TestFanout_CelebrityMode_WritesToCelebrityFeedOnly(t *testing.T) {
	authorID := uuid.New()

	userStore := &stubUserStore{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*user.User, error) {
			return &user.User{FollowersCount: 5}, nil // above pushThreshold (2)
		},
	}
	followStore := &stubFollowStore{
		getFollowerIDsFn: func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}, nil
		},
	}
	cache := &stubFeedCache{}

	f := feed.NewFanout(followStore, userStore, cache)
	f.HandlePostCreated(context.Background(), events.PostCreatedEvent{
		PostID:    uuid.New(),
		UserID:    authorID,
		CreatedAt: time.Now(),
	})

	if !cache.celebrityFeedCalled {
		t.Error("expected AddToCelebrityFeed to be called for celebrity account")
	}
	if len(cache.addPostCalled) > 0 {
		t.Errorf("expected 0 AddPost calls in celebrity mode, got %d", len(cache.addPostCalled))
	}
}

func TestFanout_AuthorNotFound_DoesNotPanic(t *testing.T) {
	userStore := &stubUserStore{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*user.User, error) {
			return nil, context.DeadlineExceeded
		},
	}
	cache := &stubFeedCache{}
	followStore := &stubFollowStore{}

	f := feed.NewFanout(followStore, userStore, cache)
	f.HandlePostCreated(context.Background(), events.PostCreatedEvent{
		PostID:    uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Now(),
	})
	// should log and return without panic
}

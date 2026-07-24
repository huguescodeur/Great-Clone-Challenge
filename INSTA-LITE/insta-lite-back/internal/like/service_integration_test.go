//go:build integration

package like_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/like"
	"github.com/huguescodeur/insta-lite/internal/testutil"
)

func TestLikeService_CreateLike_IncrementsRedisCounter(t *testing.T) {
	db := testutil.TestDB(t)
	rdb := testutil.TestRedis(t)

	authorID := testutil.InsertTestUser(t, db)
	postID := testutil.InsertTestPost(t, db, authorID)
	likerID := testutil.InsertTestUser(t, db)

	counterKey := "likes_count:" + postID.String()
	testutil.CleanRedisKey(t, rdb, counterKey)

	svc := like.NewServiceLike(
		like.NewLikeStore(db),
		&stubPostUpdater{},
		like.NewRedisLikeCounter(rdb),
		&stubAuthorGetter{
			getAuthorIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
				return authorID, nil
			},
		},
		newBus(),
		db,
	)

	_, err := svc.CreateLike(context.Background(), like.Like{
		PostID:       postID,
		UserID:       likerID,
		ReactionType: like.ReactionLike,
	})
	if err != nil {
		t.Fatalf("CreateLike: %v", err)
	}

	val, err := rdb.Get(context.Background(), counterKey).Int64()
	if err != nil {
		t.Fatalf("Redis GET %s: %v", counterKey, err)
	}
	if val != 1 {
		t.Errorf("want counter=1, got %d", val)
	}
}

func TestLikeService_Delete_DecrementsRedisCounter(t *testing.T) {
	db := testutil.TestDB(t)
	rdb := testutil.TestRedis(t)

	authorID := testutil.InsertTestUser(t, db)
	postID := testutil.InsertTestPost(t, db, authorID)
	likerID := testutil.InsertTestUser(t, db)

	counterKey := "likes_count:" + postID.String()
	testutil.CleanRedisKey(t, rdb, counterKey)

	svc := like.NewServiceLike(
		like.NewLikeStore(db),
		&stubPostUpdater{},
		like.NewRedisLikeCounter(rdb),
		&stubAuthorGetter{
			getAuthorIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
				return authorID, nil
			},
		},
		newBus(),
		db,
	)

	if _, err := svc.CreateLike(context.Background(), like.Like{
		PostID:       postID,
		UserID:       likerID,
		ReactionType: like.ReactionLike,
	}); err != nil {
		t.Fatalf("CreateLike setup: %v", err)
	}

	before, _ := rdb.Get(context.Background(), counterKey).Int64()
	if before != 1 {
		t.Fatalf("setup: want counter=1, got %d", before)
	}

	if err := svc.Delete(context.Background(), postID, likerID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	after, err := rdb.Get(context.Background(), counterKey).Int64()
	if err != nil {
		t.Fatalf("Redis GET after delete: %v", err)
	}
	if after != 0 {
		t.Errorf("want counter=0 after delete, got %d", after)
	}
}

func TestLikeService_SelfLike_NoEventPublishedOnBus(t *testing.T) {
	db := testutil.TestDB(t)
	rdb := testutil.TestRedis(t)

	userID := testutil.InsertTestUser(t, db)
	postID := testutil.InsertTestPost(t, db, userID)

	counterKey := "likes_count:" + postID.String()
	testutil.CleanRedisKey(t, rdb, counterKey)

	bus := newBus()
	svc := like.NewServiceLike(
		like.NewLikeStore(db),
		&stubPostUpdater{},
		like.NewRedisLikeCounter(rdb),
		&stubAuthorGetter{
			getAuthorIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
				return userID, nil // author == liker
			},
		},
		bus,
		db,
	)

	if _, err := svc.CreateLike(context.Background(), like.Like{
		PostID:       postID,
		UserID:       userID,
		ReactionType: like.ReactionLike,
	}); err != nil {
		t.Fatalf("CreateLike: %v", err)
	}

	select {
	case <-bus.Events():
		t.Error("expected no event for self-like, but got one")
	default:
		// correct: bus is empty
	}
}

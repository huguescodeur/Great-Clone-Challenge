//go:build integration

package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/feed"
	"github.com/huguescodeur/insta-lite/internal/post"
	"github.com/huguescodeur/insta-lite/internal/testutil"
)

func TestFeedCache_AddPost_GetFeed_ReturnsMostRecentFirst(t *testing.T) {
	rdb := testutil.TestRedis(t)
	cache := feed.NewRedisFeedCache(rdb)

	userID := uuid.New()
	testutil.CleanRedisKey(t, rdb, "feed:"+userID.String())

	now := time.Now().Truncate(time.Millisecond)
	postOlder := uuid.New()
	postNewer := uuid.New()

	if err := cache.AddPost(context.Background(), userID, postOlder, now.Add(-time.Second)); err != nil {
		t.Fatalf("AddPost older: %v", err)
	}
	if err := cache.AddPost(context.Background(), userID, postNewer, now); err != nil {
		t.Fatalf("AddPost newer: %v", err)
	}

	entries, err := cache.GetFeed(context.Background(), userID, nil, 10)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("want at least 2 entries, got %d", len(entries))
	}
	if entries[0].PostID != postNewer {
		t.Errorf("want newest first (postNewer), got %v", entries[0].PostID)
	}
	if entries[1].PostID != postOlder {
		t.Errorf("want older second (postOlder), got %v", entries[1].PostID)
	}
}

func TestFeedCache_GetFeed_CursorPagination(t *testing.T) {
	rdb := testutil.TestRedis(t)
	cache := feed.NewRedisFeedCache(rdb)

	userID := uuid.New()
	testutil.CleanRedisKey(t, rdb, "feed:"+userID.String())

	now := time.Now().Truncate(time.Millisecond)
	postIDs := make([]uuid.UUID, 4)
	for i := range postIDs {
		postIDs[i] = uuid.New()
		if err := cache.AddPost(context.Background(), userID, postIDs[i], now.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatalf("AddPost %d: %v", i, err)
		}
	}

	// Page 1: newest 2 posts (indices 3 and 2)
	page1, err := cache.GetFeed(context.Background(), userID, nil, 2)
	if err != nil {
		t.Fatalf("GetFeed page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("want 2 in page1, got %d", len(page1))
	}

	// Page 2: using cursor of last entry from page1
	cursor := &post.PostCursor{
		CreatedAt: page1[len(page1)-1].CreatedAt,
		PostID:    page1[len(page1)-1].PostID,
	}
	page2, err := cache.GetFeed(context.Background(), userID, cursor, 2)
	if err != nil {
		t.Fatalf("GetFeed page2: %v", err)
	}
	if len(page2) < 1 {
		t.Fatalf("want at least 1 entry in page2, got %d", len(page2))
	}
	// Entries in page2 must be strictly older than page1's last entry
	if !page2[0].CreatedAt.Before(page1[len(page1)-1].CreatedAt) {
		t.Errorf("page2 entry should be older than page1 last entry")
	}
}

func TestFeedCache_AddToCelebrityFeed_GetCelebrityFeed(t *testing.T) {
	rdb := testutil.TestRedis(t)
	cache := feed.NewRedisFeedCache(rdb)

	authorID := uuid.New()
	testutil.CleanRedisKey(t, rdb, "celebrity_feed:"+authorID.String())

	now := time.Now().Truncate(time.Millisecond)
	post1 := uuid.New()
	post2 := uuid.New()

	if err := cache.AddToCelebrityFeed(context.Background(), authorID, post1, now.Add(-time.Second)); err != nil {
		t.Fatalf("AddToCelebrityFeed post1: %v", err)
	}
	if err := cache.AddToCelebrityFeed(context.Background(), authorID, post2, now); err != nil {
		t.Fatalf("AddToCelebrityFeed post2: %v", err)
	}

	entries, err := cache.GetCelebrityFeed(context.Background(), authorID, nil, 10)
	if err != nil {
		t.Fatalf("GetCelebrityFeed: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("want at least 2 celebrity feed entries, got %d", len(entries))
	}
	if entries[0].PostID != post2 {
		t.Errorf("want newest (post2) first, got %v", entries[0].PostID)
	}
	if entries[1].PostID != post1 {
		t.Errorf("want older (post1) second, got %v", entries[1].PostID)
	}
}

func TestFeedCache_WarmCache_PopulatesUserFeed(t *testing.T) {
	rdb := testutil.TestRedis(t)
	cache := feed.NewRedisFeedCache(rdb)

	userID := uuid.New()
	testutil.CleanRedisKey(t, rdb, "feed:"+userID.String())

	now := time.Now().Truncate(time.Millisecond)
	posts := []*post.Post{
		{PostID: uuid.New(), CreatedAt: now.Add(-2 * time.Second)},
		{PostID: uuid.New(), CreatedAt: now.Add(-time.Second)},
		{PostID: uuid.New(), CreatedAt: now},
	}

	if err := cache.WarmCache(context.Background(), userID, posts); err != nil {
		t.Fatalf("WarmCache: %v", err)
	}

	entries, err := cache.GetFeed(context.Background(), userID, nil, 10)
	if err != nil {
		t.Fatalf("GetFeed after WarmCache: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("want 3 entries after WarmCache, got %d", len(entries))
	}
	// Newest first
	if entries[0].PostID != posts[2].PostID {
		t.Errorf("want newest post (index 2) first, got %v", entries[0].PostID)
	}
}

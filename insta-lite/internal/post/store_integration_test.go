//go:build integration

package post_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/post"
	"github.com/huguescodeur/insta-lite/internal/testutil"
)

func TestPostStore_Create_PersistsToDB(t *testing.T) {
	db := testutil.TestDB(t)
	store := post.NewPostStore(db)

	userID := testutil.InsertTestUser(t, db)
	postID := uuid.New()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	p, err := store.Create(context.Background(), tx, &post.Post{
		PostID:  postID,
		UserID:  userID,
		Content: "integration test content",
	})
	if err != nil {
		tx.Rollback()
		t.Fatalf("Create: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM posts WHERE post_id = $1", postID)
	})

	if p.PostID != postID {
		t.Errorf("want postID %s, got %s", postID, p.PostID)
	}
	if p.UserID != userID {
		t.Errorf("want userID %s, got %s", userID, p.UserID)
	}
	if p.Content != "integration test content" {
		t.Errorf("want content 'integration test content', got %q", p.Content)
	}
	if p.CreatedAt.IsZero() {
		t.Error("want non-zero CreatedAt")
	}
}

func TestPostStore_GetAuthorID_ReturnsCorrectUser(t *testing.T) {
	db := testutil.TestDB(t)
	store := post.NewPostStore(db)

	userID := testutil.InsertTestUser(t, db)
	postID := testutil.InsertTestPost(t, db, userID)

	got, err := store.GetAuthorID(context.Background(), postID)
	if err != nil {
		t.Fatalf("GetAuthorID: %v", err)
	}
	if got != userID {
		t.Errorf("want authorID %s, got %s", userID, got)
	}
}

func TestPostStore_GetAll_ReturnsMostRecentFirst(t *testing.T) {
	db := testutil.TestDB(t)
	store := post.NewPostStore(db)

	userID := testutil.InsertTestUser(t, db)

	for i := 0; i < 3; i++ {
		testutil.InsertTestPost(t, db, userID)
	}

	posts, err := store.GetAll(context.Background(), nil, 10)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(posts) < 3 {
		t.Fatalf("want at least 3 posts, got %d", len(posts))
	}

	for i := 1; i < len(posts); i++ {
		if posts[i].CreatedAt.After(posts[i-1].CreatedAt) {
			t.Errorf("posts[%d].CreatedAt (%v) is newer than posts[%d].CreatedAt (%v) — not sorted desc",
				i, posts[i].CreatedAt, i-1, posts[i-1].CreatedAt)
		}
	}
}

func TestPostStore_GetAllByID_OnlyReturnsUserPosts(t *testing.T) {
	db := testutil.TestDB(t)
	store := post.NewPostStore(db)

	userA := testutil.InsertTestUser(t, db)
	userB := testutil.InsertTestUser(t, db)

	testutil.InsertTestPost(t, db, userA)
	testutil.InsertTestPost(t, db, userA)
	testutil.InsertTestPost(t, db, userB)

	posts, err := store.GetAllByID(context.Background(), userA, nil, 10)
	if err != nil {
		t.Fatalf("GetAllByID: %v", err)
	}
	if len(posts) < 2 {
		t.Fatalf("want at least 2 posts for userA, got %d", len(posts))
	}
	for _, p := range posts {
		if p.UserID != userA {
			t.Errorf("want all posts owned by userA, got post owned by %s", p.UserID)
		}
	}
}

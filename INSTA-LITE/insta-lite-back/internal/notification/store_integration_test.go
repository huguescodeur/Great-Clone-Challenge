//go:build integration

package notification_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/notification"
	"github.com/huguescodeur/insta-lite/internal/testutil"
)

func TestNotificationStore_Create(t *testing.T) {
	db := testutil.TestDB(t)
	store := notification.NewNotificationStore(db)

	recipient := testutil.InsertTestUser(t, db)
	actor := testutil.InsertTestUser(t, db)
	postID := uuid.New()

	n, err := store.Create(context.Background(), &notification.Notification{
		RecipientID: recipient,
		ActorID:     actor,
		Type:        notification.TypeLikePost,
		EntityID:    &postID,
	})

	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if n.NotificationID == uuid.Nil {
		t.Error("expected non-nil NotificationID")
	}
	if n.Type != notification.TypeLikePost {
		t.Errorf("want type %s, got %s", notification.TypeLikePost, n.Type)
	}
	if n.Read {
		t.Error("expected read=false on creation")
	}
}

func TestNotificationStore_GetAllByRecipient_OrderedByCreatedAtDesc(t *testing.T) {
	db := testutil.TestDB(t)
	store := notification.NewNotificationStore(db)

	recipient := testutil.InsertTestUser(t, db)
	actor := testutil.InsertTestUser(t, db)

	types := []notification.NotificationType{
		notification.TypeLikePost,
		notification.TypeCommentPost,
		notification.TypeNewFollower,
	}
	for _, typ := range types {
		_, err := store.Create(context.Background(), &notification.Notification{
			RecipientID: recipient,
			ActorID:     actor,
			Type:        typ,
		})
		if err != nil {
			t.Fatalf("Create %s: %v", typ, err)
		}
		time.Sleep(2 * time.Millisecond) // ensure distinct created_at
	}

	got, err := store.GetAllByRecipient(context.Background(), recipient, nil, 10)
	if err != nil {
		t.Fatalf("GetAllByRecipient: %v", err)
	}

	if len(got) < 3 {
		t.Fatalf("expected at least 3 notifications, got %d", len(got))
	}

	// Most recent (new_follower) must be first
	if got[0].Type != notification.TypeNewFollower {
		t.Errorf("expected newest first (new_follower), got %s", got[0].Type)
	}
}

func TestNotificationStore_CountUnread(t *testing.T) {
	db := testutil.TestDB(t)
	store := notification.NewNotificationStore(db)

	recipient := testutil.InsertTestUser(t, db)
	actor := testutil.InsertTestUser(t, db)

	for i := 0; i < 3; i++ {
		_, err := store.Create(context.Background(), &notification.Notification{
			RecipientID: recipient,
			ActorID:     actor,
			Type:        notification.TypeLikePost,
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	count, err := store.CountUnread(context.Background(), recipient)
	if err != nil {
		t.Fatalf("CountUnread: %v", err)
	}
	if count < 3 {
		t.Errorf("expected at least 3 unread, got %d", count)
	}
}

func TestNotificationStore_MarkAsRead(t *testing.T) {
	db := testutil.TestDB(t)
	store := notification.NewNotificationStore(db)

	recipient := testutil.InsertTestUser(t, db)
	actor := testutil.InsertTestUser(t, db)

	n, err := store.Create(context.Background(), &notification.Notification{
		RecipientID: recipient,
		ActorID:     actor,
		Type:        notification.TypeNewFollower,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := store.MarkAsRead(context.Background(), recipient, n.NotificationID); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}

	// Verify read=true
	rows, err := store.GetAllByRecipient(context.Background(), recipient, nil, 10)
	if err != nil {
		t.Fatalf("GetAllByRecipient: %v", err)
	}
	for _, row := range rows {
		if row.NotificationID == n.NotificationID && !row.Read {
			t.Error("expected notification to be marked as read")
		}
	}
}

func TestNotificationStore_MarkAsRead_WrongRecipient_ReturnsError(t *testing.T) {
	db := testutil.TestDB(t)
	store := notification.NewNotificationStore(db)

	recipient := testutil.InsertTestUser(t, db)
	actor := testutil.InsertTestUser(t, db)
	otherUser := testutil.InsertTestUser(t, db)

	n, err := store.Create(context.Background(), &notification.Notification{
		RecipientID: recipient,
		ActorID:     actor,
		Type:        notification.TypeLikePost,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// other user tries to mark recipient's notification as read
	err = store.MarkAsRead(context.Background(), otherUser, n.NotificationID)
	if err == nil {
		t.Error("expected error when wrong recipient tries to mark as read")
	}
}

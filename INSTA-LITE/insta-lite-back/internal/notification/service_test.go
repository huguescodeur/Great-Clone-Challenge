package notification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/events"
	"github.com/huguescodeur/insta-lite/internal/notification"
)

// --- stub store ---

type stubStore struct {
	createFn           func(ctx context.Context, n *notification.Notification) (*notification.Notification, error)
	getAllByRecipientFn func(ctx context.Context, recipientID uuid.UUID, cursor *notification.NotificationCursor, limit int) ([]*notification.Notification, error)
	countUnreadFn      func(ctx context.Context, recipientID uuid.UUID) (int64, error)
	markAsReadFn       func(ctx context.Context, recipientID, notificationID uuid.UUID) error
}

func (s *stubStore) Create(ctx context.Context, n *notification.Notification) (*notification.Notification, error) {
	return s.createFn(ctx, n)
}

func (s *stubStore) GetAllByRecipient(ctx context.Context, recipientID uuid.UUID, cursor *notification.NotificationCursor, limit int) ([]*notification.Notification, error) {
	return s.getAllByRecipientFn(ctx, recipientID, cursor, limit)
}

func (s *stubStore) CountUnread(ctx context.Context, recipientID uuid.UUID) (int64, error) {
	return s.countUnreadFn(ctx, recipientID)
}

func (s *stubStore) MarkAsRead(ctx context.Context, recipientID, notificationID uuid.UUID) error {
	return s.markAsReadFn(ctx, recipientID, notificationID)
}

// --- helpers ---

func newService(store *stubStore) *notification.NotificationService {
	return notification.NewServiceNotification(store)
}

// --- Handle* tests ---

func TestHandlePostLiked_CreatesNotification(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	actorID := uuid.New()

	var captured *notification.Notification
	store := &stubStore{
		createFn: func(_ context.Context, n *notification.Notification) (*notification.Notification, error) {
			captured = n
			n.NotificationID = uuid.New()
			return n, nil
		},
	}

	svc := newService(store)
	svc.HandlePostLiked(context.Background(), events.PostLikedEvent{
		PostID:       postID,
		PostAuthorID: authorID,
		ActorID:      actorID,
	})

	if captured == nil {
		t.Fatal("expected store.Create to be called")
	}
	if captured.Type != notification.TypeLikePost {
		t.Errorf("want type %s, got %s", notification.TypeLikePost, captured.Type)
	}
	if captured.RecipientID != authorID {
		t.Errorf("want recipientID %s, got %s", authorID, captured.RecipientID)
	}
	if captured.ActorID != actorID {
		t.Errorf("want actorID %s, got %s", actorID, captured.ActorID)
	}
	if captured.EntityID == nil || *captured.EntityID != postID {
		t.Errorf("want entityID %s, got %v", postID, captured.EntityID)
	}
}

func TestHandleCommentCreated_CreatesNotification(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	actorID := uuid.New()

	var captured *notification.Notification
	store := &stubStore{
		createFn: func(_ context.Context, n *notification.Notification) (*notification.Notification, error) {
			captured = n
			n.NotificationID = uuid.New()
			return n, nil
		},
	}

	newService(store).HandleCommentCreated(context.Background(), events.CommentCreatedEvent{
		PostID:       postID,
		PostAuthorID: authorID,
		ActorID:      actorID,
	})

	if captured == nil {
		t.Fatal("expected store.Create to be called")
	}
	if captured.Type != notification.TypeCommentPost {
		t.Errorf("want type %s, got %s", notification.TypeCommentPost, captured.Type)
	}
	if captured.RecipientID != authorID {
		t.Errorf("want recipientID %s, got %s", authorID, captured.RecipientID)
	}
	if captured.EntityID == nil || *captured.EntityID != postID {
		t.Errorf("want entityID %s, got %v", postID, captured.EntityID)
	}
}

func TestHandleCommentLiked_CreatesNotification(t *testing.T) {
	commentID := uuid.New()
	commentAuthorID := uuid.New()
	actorID := uuid.New()

	var captured *notification.Notification
	store := &stubStore{
		createFn: func(_ context.Context, n *notification.Notification) (*notification.Notification, error) {
			captured = n
			n.NotificationID = uuid.New()
			return n, nil
		},
	}

	newService(store).HandleCommentLiked(context.Background(), events.CommentLikedEvent{
		CommentID:       commentID,
		CommentAuthorID: commentAuthorID,
		ActorID:         actorID,
	})

	if captured == nil {
		t.Fatal("expected store.Create to be called")
	}
	if captured.Type != notification.TypeLikeComment {
		t.Errorf("want type %s, got %s", notification.TypeLikeComment, captured.Type)
	}
	if captured.RecipientID != commentAuthorID {
		t.Errorf("want recipientID %s, got %s", commentAuthorID, captured.RecipientID)
	}
	if captured.EntityID == nil || *captured.EntityID != commentID {
		t.Errorf("want entityID %s, got %v", commentID, captured.EntityID)
	}
}

func TestHandleUserFollowed_CreatesNotification_EntityIDIsNil(t *testing.T) {
	followeeID := uuid.New()
	actorID := uuid.New()

	var captured *notification.Notification
	store := &stubStore{
		createFn: func(_ context.Context, n *notification.Notification) (*notification.Notification, error) {
			captured = n
			n.NotificationID = uuid.New()
			return n, nil
		},
	}

	newService(store).HandleUserFollowed(context.Background(), events.UserFollowedEvent{
		FolloweeID: followeeID,
		ActorID:    actorID,
	})

	if captured == nil {
		t.Fatal("expected store.Create to be called")
	}
	if captured.Type != notification.TypeNewFollower {
		t.Errorf("want type %s, got %s", notification.TypeNewFollower, captured.Type)
	}
	if captured.EntityID != nil {
		t.Errorf("want nil entityID for new_follower, got %s", *captured.EntityID)
	}
}

func TestHandlePostLiked_StoreError_DoesNotPanic(t *testing.T) {
	store := &stubStore{
		createFn: func(_ context.Context, _ *notification.Notification) (*notification.Notification, error) {
			return nil, errors.New("db error")
		},
	}
	// should log the error but not panic
	newService(store).HandlePostLiked(context.Background(), events.PostLikedEvent{
		PostID:       uuid.New(),
		PostAuthorID: uuid.New(),
		ActorID:      uuid.New(),
	})
}

// --- GetNotifications tests ---

func TestGetNotifications_HasMore_AndNextCursor(t *testing.T) {
	recipientID := uuid.New()
	now := time.Now()

	notifs := make([]*notification.Notification, 11)
	for i := range notifs {
		notifs[i] = &notification.Notification{
			NotificationID: uuid.New(),
			RecipientID:    recipientID,
			ActorID:        uuid.New(),
			Type:           notification.TypeLikePost,
			Read:           false,
			CreatedAt:      now.Add(-time.Duration(i) * time.Minute),
		}
	}

	store := &stubStore{
		getAllByRecipientFn: func(_ context.Context, _ uuid.UUID, _ *notification.NotificationCursor, limit int) ([]*notification.Notification, error) {
			return notifs[:limit], nil
		},
		countUnreadFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			return 5, nil
		},
	}

	res, err := newService(store).GetNotifications(context.Background(), recipientID, "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasMore {
		t.Error("expected hasMore=true")
	}
	if res.NextCursor == "" {
		t.Error("expected non-empty nextCursor")
	}
	if len(res.Data) != 10 {
		t.Errorf("expected 10 items in data, got %d", len(res.Data))
	}
	if res.UnreadCount != 5 {
		t.Errorf("expected unreadCount=5, got %d", res.UnreadCount)
	}
}

func TestGetNotifications_NoMore(t *testing.T) {
	recipientID := uuid.New()

	store := &stubStore{
		getAllByRecipientFn: func(_ context.Context, _ uuid.UUID, _ *notification.NotificationCursor, limit int) ([]*notification.Notification, error) {
			return []*notification.Notification{
				{NotificationID: uuid.New(), CreatedAt: time.Now(), Type: notification.TypeLikePost},
			}, nil
		},
		countUnreadFn: func(_ context.Context, _ uuid.UUID) (int64, error) { return 1, nil },
	}

	res, err := newService(store).GetNotifications(context.Background(), recipientID, "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.HasMore {
		t.Error("expected hasMore=false")
	}
	if res.NextCursor != "" {
		t.Errorf("expected empty nextCursor, got %s", res.NextCursor)
	}
}

// --- MarkAsRead tests ---

func TestMarkAsRead_NotFound(t *testing.T) {
	store := &stubStore{
		markAsReadFn: func(_ context.Context, _, _ uuid.UUID) error {
			return notification.ErrNotificationNotFound
		},
	}

	err := newService(store).MarkAsRead(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, notification.ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound, got %v", err)
	}
}

func TestMarkAsRead_Success(t *testing.T) {
	called := false
	store := &stubStore{
		markAsReadFn: func(_ context.Context, _, _ uuid.UUID) error {
			called = true
			return nil
		},
	}

	err := newService(store).MarkAsRead(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected store.MarkAsRead to be called")
	}
}

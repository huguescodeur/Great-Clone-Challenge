package notification

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/events"
)

type NotificationService struct {
	store NotificationStore
}

func NewServiceNotification(store NotificationStore) *NotificationService {
	return &NotificationService{store: store}
}

func (s *NotificationService) GetNotifications(ctx context.Context, recipientID uuid.UUID, cursorStr string, limit int) (*PaginatedNotificationResponse, error) {
	cursor, err := DecodeCursor(cursorStr)
	if err != nil {
		return nil, err
	}

	notifications, err := s.store.GetAllByRecipient(ctx, recipientID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(notifications) > limit
	if hasMore {
		notifications = notifications[:limit]
	}

	unreadCount, err := s.store.CountUnread(ctx, recipientID)
	if err != nil {
		log.Printf("[notification] échec comptage non-lues recipientID=%s: %v", recipientID, err)
	}

	nextCursor := ""
	if hasMore && len(notifications) > 0 {
		last := notifications[len(notifications)-1]
		nextCursor = EncodeCursor(last.CreatedAt, last.NotificationID)
	}

	return &PaginatedNotificationResponse{
		Data:        notifications,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
		UnreadCount: unreadCount,
	}, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, recipientID, notificationID uuid.UUID) error {
	return s.store.MarkAsRead(ctx, recipientID, notificationID)
}

func (s *NotificationService) HandlePostLiked(ctx context.Context, e events.PostLikedEvent) {
	_, err := s.store.Create(ctx, &Notification{
		RecipientID: e.PostAuthorID,
		ActorID:     e.ActorID,
		Type:        TypeLikePost,
		EntityID:    &e.PostID,
	})
	if err != nil {
		log.Printf("[notification] échec création like_post postID=%s: %v", e.PostID, err)
	}
}

func (s *NotificationService) HandleCommentCreated(ctx context.Context, e events.CommentCreatedEvent) {
	_, err := s.store.Create(ctx, &Notification{
		RecipientID: e.PostAuthorID,
		ActorID:     e.ActorID,
		Type:        TypeCommentPost,
		EntityID:    &e.PostID,
	})
	if err != nil {
		log.Printf("[notification] échec création comment_post postID=%s: %v", e.PostID, err)
	}
}

func (s *NotificationService) HandleCommentLiked(ctx context.Context, e events.CommentLikedEvent) {
	_, err := s.store.Create(ctx, &Notification{
		RecipientID: e.CommentAuthorID,
		ActorID:     e.ActorID,
		Type:        TypeLikeComment,
		EntityID:    &e.CommentID,
	})
	if err != nil {
		log.Printf("[notification] échec création like_comment commentID=%s: %v", e.CommentID, err)
	}
}

func (s *NotificationService) HandleUserFollowed(ctx context.Context, e events.UserFollowedEvent) {
	_, err := s.store.Create(ctx, &Notification{
		RecipientID: e.FolloweeID,
		ActorID:     e.ActorID,
		Type:        TypeNewFollower,
		EntityID:    nil,
	})
	if err != nil {
		log.Printf("[notification] échec création new_follower actorID=%s: %v", e.ActorID, err)
	}
}

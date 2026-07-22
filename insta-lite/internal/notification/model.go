package notification

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	TypeLikePost    NotificationType = "like_post"
	TypeCommentPost NotificationType = "comment_post"
	TypeLikeComment NotificationType = "like_comment"
	TypeNewFollower NotificationType = "new_follower"
)

type Notification struct {
	NotificationID uuid.UUID        `json:"notificationId"`
	RecipientID    uuid.UUID        `json:"recipientId"`
	ActorID        uuid.UUID        `json:"actorId"`
	Type           NotificationType `json:"type"`
	EntityID       *uuid.UUID       `json:"entityId,omitempty"`
	Read           bool             `json:"read"`
	CreatedAt      time.Time        `json:"createdAt"`
}

type NotificationCursor struct {
	CreatedAt      time.Time
	NotificationID uuid.UUID
}

package events

import (
	"time"

	"github.com/google/uuid"
)

type PostCreatedEvent struct {
	PostID    uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
}

type PostLikedEvent struct {
	PostID       uuid.UUID
	PostAuthorID uuid.UUID
	ActorID      uuid.UUID // celui qui a liké
	CreatedAt    time.Time
}

type CommentCreatedEvent struct {
	CommentID    uuid.UUID
	PostID       uuid.UUID
	PostAuthorID uuid.UUID
	ActorID      uuid.UUID // celui qui a commenté
	CreatedAt    time.Time
}

type CommentLikedEvent struct {
	CommentID       uuid.UUID
	CommentAuthorID uuid.UUID
	ActorID         uuid.UUID
	CreatedAt       time.Time
}

type UserFollowedEvent struct {
	FolloweeID uuid.UUID // celui qui est suivi (reçoit la notif)
	ActorID    uuid.UUID // celui qui suit
	CreatedAt  time.Time
}

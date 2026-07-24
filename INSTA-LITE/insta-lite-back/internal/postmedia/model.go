package postmedia

import (
	"time"

	"github.com/google/uuid"
)

type PostMedia struct {
	MediaID   uuid.UUID `json:"mediaID"`
	PostID    uuid.UUID `json:"postID"`
	URL       string    `json:"url" validate:"required,url"`
	Type      string    `json:"type" validate:"required,oneof=image video"`
	CreatedAt time.Time `json:"createdAt"`
}

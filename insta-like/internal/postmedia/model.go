package postmedia

import (
	"time"

	"github.com/google/uuid"
)

type PostMedia struct {
	MediaID   uuid.UUID `json:"mediaID"`
	PostID    uuid.UUID `json:"postID"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
}

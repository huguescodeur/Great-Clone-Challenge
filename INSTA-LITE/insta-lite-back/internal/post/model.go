package post

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	PostID   uuid.UUID `json:"postID"`
	UserID   uuid.UUID `json:"userID"`
	Username string    `json:"username"`

	Content       string `json:"content"`
	LikesCount    int64  `json:"likesCount"`
	CommentsCount int64  `json:"commentsCount"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PostCursor struct {
	CreatedAt time.Time
	PostID    uuid.UUID
}

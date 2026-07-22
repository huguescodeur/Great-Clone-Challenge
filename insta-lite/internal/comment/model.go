package comment

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	CommentID       uuid.UUID  `json:"commentId"`
	PostID          uuid.UUID  `json:"postId"`
	UserID          uuid.UUID  `json:"userId"`
	ParentCommentID *uuid.UUID `json:"parentCommentId,omitempty"`
	Content         string     `json:"content"`
	LikesCount      int64      `json:"likesCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type CommentCursor struct {
	CreatedAt time.Time
	CommentID uuid.UUID
}

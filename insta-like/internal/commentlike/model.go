package commentlike

import (
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-like/internal/like"
)

type CommentLike struct {
	CommentID    uuid.UUID         `json:"commentId"`
	UserID       uuid.UUID         `json:"userId"`
	ReactionType like.ReactionType `json:"reactionType"`
	Created_At   time.Time         `json:"createdAt"`
	Updated_At   time.Time         `json:"updatedAt"`
}

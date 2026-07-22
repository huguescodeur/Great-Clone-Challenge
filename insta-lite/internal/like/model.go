package like

import (
	"time"

	"github.com/google/uuid"
)

type ReactionType string

const (
	ReactionLike  ReactionType = "like"
	ReactionLove  ReactionType = "love"
	ReactionLaugh ReactionType = "laugh"
)

type Like struct {
	PostID       uuid.UUID    `json:"postID"`
	UserID       uuid.UUID    `json:"userID"`
	ReactionType ReactionType `json:"reactionType"`
	Created_At   time.Time    `json:"created_at"`
	Updated_At   time.Time    `json:"updated_at"`
}

func (r ReactionType) IsValid() bool {
	switch r {
	case ReactionLike, ReactionLove, ReactionLaugh:
		return true
	default:
		return false
	}
}

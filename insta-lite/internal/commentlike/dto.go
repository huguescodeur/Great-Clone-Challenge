package commentlike

import "github.com/huguescodeur/insta-lite/internal/like"

type reactionRequest struct {
	ReactionType like.ReactionType `json:"reactionType" validate:"required"`
}

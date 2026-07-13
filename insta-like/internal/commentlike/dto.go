package commentlike

import "github.com/huguescodeur/insta-like/internal/like"

type reactionRequest struct {
	ReactionType like.ReactionType `json:"reactionType" validate:"required"`
}

package like

type LikeRequest struct {
	ReactionType ReactionType `json:"reactionType" validate:"required"`
}

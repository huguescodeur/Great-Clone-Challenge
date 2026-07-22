package comment

import (
	"github.com/google/uuid"
)

type PaginatedCommentResponse struct {
	Data       []*Comment `json:"data"`
	NextCursor string     `json:"nextCursor"`
	HasMore    bool       `json:"hasMore"`
}

type CommentCreateRequest struct {
	ParentCommentID *uuid.UUID `json:"parentCommentId"`
	Content         string     `json:"content" validate:"required,max=2000"`
}

type CommentUpdateRequest struct {
	Content string `json:"content" validate:"required,max=2000"`
}

type CommentDeleteRequest struct {
	CommentID uuid.UUID `json:"commentId" validate:"required"`
}

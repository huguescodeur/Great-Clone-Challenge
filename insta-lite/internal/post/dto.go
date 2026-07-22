package post

import (
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/postmedia"
)

type PostResponse struct {
	Post
	Medias []postmedia.PostMedia `json:"media"`
}

type PaginatedPostResponse struct {
	Data       []*PostResponse `json:"data"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

type PostUpdateRequest struct {
	PostID  uuid.UUID `json:"postID" validate:"required"`
	Content string    `json:"content" validate:"required"`
}

type PostDeleteRequest struct {
	PostID uuid.UUID `json:"postID" validate:"required"`
}

type PostCreateRequest struct {
	Content string                `json:"content" validate:"required"`
	Medias  []postmedia.PostMedia `json:"medias" validate:"required"`
}

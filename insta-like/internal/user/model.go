package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID        uuid.UUID `json:"userID"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	FullName      string    `json:"fullName"`
	Bio           string    `json:"bio"`
	ProfilePicURL string    `json:"profilePicURL"`

	FollowersCount int64 `json:"followersCount"`
	FollowingCount int64 `json:"followingCount"`
	PostsCount     int64 `json:"postsCount"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"-"`
}

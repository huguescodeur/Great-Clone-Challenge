package follow

import (
	"time"

	"github.com/google/uuid"
)

type FollowStatus string

const (
	StatusPending  FollowStatus = "pending"
	StatusAccepted FollowStatus = "accepted"
)

func (s FollowStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusAccepted:
		return true
	default:
		return false
	}
}

type Follow struct {
	FollowerID       uuid.UUID    `json:"followerId"`
	FolloweeID       uuid.UUID    `json:"followeeId"`
	FollowerUsername string       `json:"followerUsername,omitempty"`
	FolloweeUsername string       `json:"followeeUsername,omitempty"`
	Status           FollowStatus `json:"status"`
	CreatedAt        time.Time    `json:"createdAt"`
}

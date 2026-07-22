package feed

import (
	"context"
	"log"

	"github.com/huguescodeur/insta-lite/internal/events"
	"github.com/huguescodeur/insta-lite/internal/follow"
	"github.com/huguescodeur/insta-lite/internal/user"
)

const pushThreshold = 2

type Fanout struct {
	followStore follow.FollowStore
	userStore   user.UserStore
	cache       FeedCache
}

func NewFanout(followStore follow.FollowStore, userStore user.UserStore, cache FeedCache) *Fanout {
	return &Fanout{followStore: followStore, userStore: userStore, cache: cache}
}

func (f *Fanout) HandlePostCreated(ctx context.Context, e events.PostCreatedEvent) {
	author, err := f.userStore.GetByID(ctx, e.UserID)
	if err != nil {
		log.Printf("[fanout] erreur récupération auteur userID=%s: %v", e.UserID, err)
		return
	}

	if author.FollowersCount > pushThreshold {
		// Mode pull : un seul write, peu importe le nombre de followers.
		if err := f.cache.AddToCelebrityFeed(ctx, e.UserID, e.PostID, e.CreatedAt); err != nil {
			log.Printf("[fanout] erreur écriture celebrity feed userID=%s: %v", e.UserID, err)
			return
		}
		log.Printf("[fanout] post %s en mode pull (compte %s, %d followers)", e.PostID, e.UserID, author.FollowersCount)
		return
	}

	// Mode push : comportement existant, inchangé.
	followerIDs, err := f.followStore.GetFollowerIDs(ctx, e.UserID)
	if err != nil {
		log.Printf("[fanout] erreur récupération followers userID=%s: %v", e.UserID, err)
		return
	}

	pushed := 0
	for _, followerID := range followerIDs {
		if err := f.cache.AddPost(ctx, followerID, e.PostID, e.CreatedAt); err != nil {
			log.Printf("[fanout] erreur push followerID=%s postID=%s: %v", followerID, e.PostID, err)
			continue
		}
		pushed++
	}
	log.Printf("[fanout] post %s distribué à %d/%d followers (mode push)", e.PostID, pushed, len(followerIDs))
}

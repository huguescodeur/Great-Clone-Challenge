package feed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/follow"
	"github.com/huguescodeur/insta-lite/internal/post"
	"github.com/huguescodeur/insta-lite/internal/postmedia"
	"github.com/huguescodeur/insta-lite/internal/user"
)

type LikesCounter interface {
	GetCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type CommentsCounter interface {
	GetCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type FeedService struct {
	feedStore       FeedStore
	feedCache       FeedCache
	followStore     follow.FollowStore
	userStore       user.UserStore
	postStore       post.PostStore
	postMediaStore  postmedia.PostMediaStore
	likesCounter    LikesCounter
	commentsCounter CommentsCounter
}

func NewServiceFeed(f FeedStore, c FeedCache, fs follow.FollowStore, us user.UserStore, p post.PostStore, m postmedia.PostMediaStore, lc LikesCounter, cc CommentsCounter) *FeedService {
	return &FeedService{feedStore: f, feedCache: c, followStore: fs, userStore: us, postStore: p, postMediaStore: m, likesCounter: lc, commentsCounter: cc}
}

func (s *FeedService) GetFeed(ctx context.Context, followerID uuid.UUID, cursorStr string, limit int) (*post.PaginatedPostResponse, error) {
	cursor, err := post.DecodeCursor(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}

	entries, err := s.feedCache.GetFeed(ctx, followerID, cursor, limit+1)
	if err != nil {
		log.Printf("[feed] erreur lecture cache userID=%s: %v", followerID, err)
		entries = nil
	}

	celebrityEntries, err := s.getCelebrityEntries(ctx, followerID, cursor, limit+1)
	if err != nil {

		log.Printf("[feed] erreur lecture celebrity feeds userID=%s: %v", followerID, err)
	}

	merged := mergeEntries(entries, celebrityEntries, limit+1)

	var posts []*post.Post

	if len(merged) == 0 && cursor == nil {
		posts, err = s.feedStore.GetFeed(ctx, followerID, nil, limit+1)
		if err != nil {
			return nil, err
		}

		if len(posts) > 0 {
			go func(userID uuid.UUID, toCache []*post.Post) {
				warmCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				if err := s.feedCache.WarmCache(warmCtx, userID, toCache); err != nil {
					log.Printf("[feed] échec réchauffement cache userID=%s: %v", userID, err)
				}
			}(followerID, posts)
		}

		if len(posts) == 0 {
			posts, err = s.postStore.GetAll(ctx, nil, limit+1)
			if err != nil {
				return nil, err
			}
		}
	} else {
		hasMore := len(merged) > limit
		if hasMore {
			merged = merged[:limit]
		}

		ids := make([]uuid.UUID, 0, len(merged))
		for _, e := range merged {
			ids = append(ids, e.PostID)
		}

		fetched, err := s.postStore.GetByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}

		byID := make(map[uuid.UUID]*post.Post, len(fetched))
		for _, p := range fetched {
			byID[p.PostID] = p
		}
		for _, id := range ids {
			if p, ok := byID[id]; ok {
				posts = append(posts, p)
			}
		}
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	s.applyLiveLikesCounts(ctx, posts)
	s.applyLiveCommentsCounts(ctx, posts)

	var ids []uuid.UUID
	for _, p := range posts {
		ids = append(ids, p.PostID)
	}
	media, err := s.postMediaStore.GetByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	mediaMap := make(map[uuid.UUID][]postmedia.PostMedia)
	for _, m := range media {
		mediaMap[m.PostID] = append(mediaMap[m.PostID], *m)
	}

	postResponses := make([]*post.PostResponse, 0, len(posts))
	for _, p := range posts {
		postResponses = append(postResponses, &post.PostResponse{Post: *p, Medias: mediaMap[p.PostID]})
	}

	nextCursor := ""
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		nextCursor = post.EncodeCursor(last.CreatedAt, last.PostID)
	}

	return &post.PaginatedPostResponse{Data: postResponses, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s *FeedService) getCelebrityEntries(ctx context.Context, followerID uuid.UUID, cursor *post.PostCursor, limit int) ([]FeedEntry, error) {
	followeeIDs, err := s.followStore.GetFolloweeIDs(ctx, followerID)
	if err != nil || len(followeeIDs) == 0 {
		return nil, err
	}

	counts, err := s.userStore.GetFollowersCounts(ctx, followeeIDs)
	if err != nil {
		return nil, err
	}

	var all []FeedEntry
	for _, id := range followeeIDs {
		if counts[id] <= pushThreshold {
			continue
		}

		entries, err := s.feedCache.GetCelebrityFeed(ctx, id, cursor, limit)
		if err != nil {
			log.Printf("[feed] erreur celebrity feed authorID=%s: %v", id, err)
			continue
		}

		if len(entries) == 0 && cursor == nil {

			entries = s.backfillCelebrityFeed(ctx, id, limit)
		}

		all = append(all, entries...)
	}
	return all, nil
}

func (s *FeedService) backfillCelebrityFeed(ctx context.Context, authorID uuid.UUID, limit int) []FeedEntry {
	posts, err := s.postStore.GetAllByID(ctx, authorID, nil, limit)
	if err != nil {
		log.Printf("[feed] échec backfill celebrity authorID=%s: %v", authorID, err)
		return nil
	}
	if len(posts) == 0 {
		return nil
	}

	entries := make([]FeedEntry, 0, len(posts))
	for _, p := range posts {
		entries = append(entries, FeedEntry{PostID: p.PostID, CreatedAt: p.CreatedAt})
	}

	go func(id uuid.UUID, toCache []*post.Post) {
		warmCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		for _, p := range toCache {
			if err := s.feedCache.AddToCelebrityFeed(warmCtx, id, p.PostID, p.CreatedAt); err != nil {
				log.Printf("[feed] échec réchauffement celebrity_feed authorID=%s: %v", id, err)
				return
			}
		}
	}(authorID, posts)

	return entries
}

func mergeEntries(a, b []FeedEntry, limit int) []FeedEntry {
	result := make([]FeedEntry, 0, limit)
	i, j := 0, 0
	for len(result) < limit && (i < len(a) || j < len(b)) {
		if j >= len(b) || (i < len(a) && a[i].CreatedAt.After(b[j].CreatedAt)) {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}
	return result
}

func (s *FeedService) applyLiveLikesCounts(ctx context.Context, posts []*post.Post) {
	if len(posts) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(posts))
	for i, p := range posts {
		ids[i] = p.PostID
	}
	counts, err := s.likesCounter.GetCounts(ctx, ids)
	if err != nil {
		log.Printf("[feed] échec lecture compteurs likes Redis: %v", err)
		return
	}

	for _, p := range posts {
		if n, ok := counts[p.PostID]; ok {
			p.LikesCount = n
		}
	}
}

func (s *FeedService) applyLiveCommentsCounts(ctx context.Context, posts []*post.Post) {
	if len(posts) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(posts))
	for i, p := range posts {
		ids[i] = p.PostID
	}
	counts, err := s.commentsCounter.GetCounts(ctx, ids)
	if err != nil {
		log.Printf("[post] échec lecture compteurs comments Redis: %v", err)
		return
	}
	for _, p := range posts {
		if n, ok := counts[p.PostID]; ok {
			p.CommentsCount = n
		}
	}
}

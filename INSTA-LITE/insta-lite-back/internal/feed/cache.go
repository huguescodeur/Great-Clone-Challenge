package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/post"
	"github.com/redis/go-redis/v9"
)

type FeedEntry struct {
	PostID    uuid.UUID
	CreatedAt time.Time
}

type FeedCache interface {
	AddPost(ctx context.Context, userID, postID uuid.UUID, postCreatedAt time.Time) error
	GetFeed(ctx context.Context, userID uuid.UUID, cursor *post.PostCursor, limit int) ([]FeedEntry, error)
	WarmCache(ctx context.Context, userID uuid.UUID, posts []*post.Post) error
	AddToCelebrityFeed(ctx context.Context, authorID, postID uuid.UUID, postCreatedAt time.Time) error
	GetCelebrityFeed(ctx context.Context, authorID uuid.UUID, cursor *post.PostCursor, limit int) ([]FeedEntry, error)
}

type redisFeedCache struct {
	client  *redis.Client
	maxSize int64
}

func NewRedisFeedCache(client *redis.Client) FeedCache {
	return &redisFeedCache{client: client, maxSize: 1000}
}

func feedKey(userID uuid.UUID) string {
	return "feed:" + userID.String()
}

func (c *redisFeedCache) AddPost(ctx context.Context, userID, postID uuid.UUID, postCreatedAt time.Time) error {
	key := feedKey(userID)
	score := float64(postCreatedAt.UnixNano())

	if err := c.client.ZAdd(ctx, key, redis.Z{Score: score, Member: postID.String()}).Err(); err != nil {
		return err
	}

	return c.client.ZRemRangeByRank(ctx, key, 0, -c.maxSize-1).Err()
}

func (c *redisFeedCache) GetFeed(ctx context.Context, userID uuid.UUID, cursor *post.PostCursor, limit int) ([]FeedEntry, error) {
	key := feedKey(userID)

	max := "+inf"
	if cursor != nil {
		max = fmt.Sprintf("(%d", cursor.CreatedAt.UnixNano())
	}

	results, err := c.client.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   max,
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]FeedEntry, 0, len(results))
	for _, z := range results {
		postID, err := uuid.Parse(z.Member.(string))
		if err != nil {
			continue
		}
		entries = append(entries, FeedEntry{
			PostID:    postID,
			CreatedAt: time.Unix(0, int64(z.Score)),
		})
	}
	return entries, nil
}

func (c *redisFeedCache) WarmCache(ctx context.Context, userID uuid.UUID, posts []*post.Post) error {
	if len(posts) == 0 {
		return nil
	}

	key := feedKey(userID)
	members := make([]redis.Z, 0, len(posts))
	for _, p := range posts {
		members = append(members, redis.Z{
			Score:  float64(p.CreatedAt.UnixNano()),
			Member: p.PostID.String(),
		})
	}

	if err := c.client.ZAdd(ctx, key, members...).Err(); err != nil {
		return err
	}

	return c.client.ZRemRangeByRank(ctx, key, 0, -c.maxSize-1).Err()
}

func celebrityFeedKey(authorID uuid.UUID) string {
	return "celebrity_feed:" + authorID.String()
}

func (c *redisFeedCache) AddToCelebrityFeed(ctx context.Context, authorID, postID uuid.UUID, postCreatedAt time.Time) error {
	key := celebrityFeedKey(authorID)
	score := float64(postCreatedAt.UnixNano())

	if err := c.client.ZAdd(ctx, key, redis.Z{Score: score, Member: postID.String()}).Err(); err != nil {
		return err
	}
	return c.client.ZRemRangeByRank(ctx, key, 0, -c.maxSize-1).Err()
}

func (c *redisFeedCache) GetCelebrityFeed(ctx context.Context, authorID uuid.UUID, cursor *post.PostCursor, limit int) ([]FeedEntry, error) {
	key := celebrityFeedKey(authorID)

	max := "+inf"
	if cursor != nil {
		max = fmt.Sprintf("(%d", cursor.CreatedAt.UnixNano())
	}

	results, err := c.client.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   max,
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]FeedEntry, 0, len(results))
	for _, z := range results {
		postID, err := uuid.Parse(z.Member.(string))
		if err != nil {
			continue
		}
		entries = append(entries, FeedEntry{PostID: postID, CreatedAt: time.Unix(0, int64(z.Score))})
	}
	return entries, nil
}

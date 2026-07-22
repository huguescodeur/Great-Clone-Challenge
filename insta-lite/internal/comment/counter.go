package comment

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CommentsCounter interface {
	Increment(ctx context.Context, postID uuid.UUID) (int64, error)
	Decrement(ctx context.Context, postID uuid.UUID) (int64, error)
	GetCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type redisCommentsCounter struct {
	client *redis.Client
}

func NewRedisCommentsCounter(client *redis.Client) CommentsCounter {
	return &redisCommentsCounter{client: client}
}

func commentsCountKey(postID uuid.UUID) string {
	return "comments_count:" + postID.String()
}

func (c *redisCommentsCounter) Increment(ctx context.Context, postID uuid.UUID) (int64, error) {
	return c.client.Incr(ctx, commentsCountKey(postID)).Result()
}

func (c *redisCommentsCounter) Decrement(ctx context.Context, postID uuid.UUID) (int64, error) {
	return c.client.Decr(ctx, commentsCountKey(postID)).Result()
}

func (c *redisCommentsCounter) GetCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(postIDs) == 0 {
		return map[uuid.UUID]int64{}, nil
	}
	keys := make([]string, len(postIDs))
	for i, id := range postIDs {
		keys[i] = commentsCountKey(id)
	}
	vals, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	counts := make(map[uuid.UUID]int64, len(postIDs))
	for i, v := range vals {
		if v == nil {
			continue
		}
		n, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			continue
		}
		counts[postIDs[i]] = n
	}
	return counts, nil
}

package commentlike

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CommentLikesCounter interface {
	Increment(ctx context.Context, commentID uuid.UUID) (int64, error)
	Decrement(ctx context.Context, commentID uuid.UUID) (int64, error)
	GetCounts(ctx context.Context, commentIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type redisCommentLikesCounter struct {
	client *redis.Client
}

func NewRedisCommentLikesCounter(client *redis.Client) CommentLikesCounter {
	return &redisCommentLikesCounter{client: client}
}

func commentLikesCountKey(commentID uuid.UUID) string {
	return "comment_likes_count:" + commentID.String()
}

func (c *redisCommentLikesCounter) Increment(ctx context.Context, commentID uuid.UUID) (int64, error) {
	return c.client.Incr(ctx, commentLikesCountKey(commentID)).Result()
}

func (c *redisCommentLikesCounter) Decrement(ctx context.Context, commentID uuid.UUID) (int64, error) {
	return c.client.Decr(ctx, commentLikesCountKey(commentID)).Result()
}

func (c *redisCommentLikesCounter) GetCounts(ctx context.Context, commentIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(commentIDs) == 0 {
		return map[uuid.UUID]int64{}, nil
	}
	keys := make([]string, len(commentIDs))
	for i, id := range commentIDs {
		keys[i] = commentLikesCountKey(id)
	}
	vals, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	counts := make(map[uuid.UUID]int64, len(commentIDs))
	for i, v := range vals {
		if v == nil {
			continue
		}
		n, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			continue
		}
		counts[commentIDs[i]] = n
	}
	return counts, nil
}

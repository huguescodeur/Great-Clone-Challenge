package like

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type LikeCounter interface {
	Increment(ctx context.Context, postID uuid.UUID) (int64, error)
	Decrement(ctx context.Context, postID uuid.UUID) (int64, error)
	GetCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type redisLikeCounter struct {
	client *redis.Client
}

func NewRedisLikeCounter(client *redis.Client) LikeCounter {
	return &redisLikeCounter{client: client}
}

func likesCountKey(postID uuid.UUID) string {
	return "likes_count:" + postID.String()
}

func (c *redisLikeCounter) Increment(ctx context.Context, postID uuid.UUID) (int64, error) {
	return c.client.Incr(ctx, likesCountKey(postID)).Result()
}

func (c *redisLikeCounter) Decrement(ctx context.Context, postID uuid.UUID) (int64, error) {
	return c.client.Decr(ctx, likesCountKey(postID)).Result()
}

func (c *redisLikeCounter) GetCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(postIDs) == 0 {
		return map[uuid.UUID]int64{}, nil
	}

	keys := make([]string, len(postIDs))
	for i, id := range postIDs {
		keys[i] = likesCountKey(id)
	}

	vals, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(postIDs))
	for i, v := range vals {
		if v == nil {
			// Clé absente : on ne l'ajoute pas à la map, ce qui signale
			// à l'appelant "pas de donnée Redis" plutôt que "0 like".
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

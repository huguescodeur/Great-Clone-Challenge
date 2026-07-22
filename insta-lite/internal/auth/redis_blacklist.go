package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBlacklist struct {
	client *redis.Client
}

func NewRedisBlacklist(client *redis.Client) *RedisBlacklist {
	return &RedisBlacklist{
		client: client,
	}
}

func (r *RedisBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := "blacklist:" + token

	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val > 0, nil
}

func (r *RedisBlacklist) Blacklist(ctx context.Context, token string, expiration time.Duration) error {
	key := "blacklist:" + token

	return r.client.Set(ctx, key, "", expiration).Err()
}

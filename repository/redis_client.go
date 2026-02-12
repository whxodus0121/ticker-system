package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

// Lock: SetNX를 이용해 열쇠를 획득 시도
func (r *RedisRepository) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, "locked", expiration).Result()
}

// Unlock: 열쇠 반납
func (r *RedisRepository) Unlock(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

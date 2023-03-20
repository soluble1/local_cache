package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisCache struct {
	client redis.Cmdable
}

func NewRedisCacheV1(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return NewRedisCache(client)
}

func NewRedisCache(c redis.Cmdable) *RedisCache {
	return &RedisCache{
		client: c,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	_, err := r.client.Set(ctx, key, val, expiration).Result()
	return err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}

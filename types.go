package cache

import (
	"context"
	"time"
)

// Cache context 在本地缓存中可能没用，但在接入redis的时候就有用了
type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

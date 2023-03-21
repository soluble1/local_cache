package cache

import (
	"context"
	"time"
)

type WriteThroughCache struct {
	Cache
	// 写DB方法
	StoreFunc func(ctx context.Context, key string, val any) error
}

// Set 先更新DB
func (w *WriteThroughCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 1.先写DB
	err := w.StoreFunc(ctx, key, val)
	// 1.1 写DB失败
	if err != nil {
		return err
	}
	// 1.2 写DB成功 再写缓存
	return w.Cache.Set(ctx, key, val, expiration)
}

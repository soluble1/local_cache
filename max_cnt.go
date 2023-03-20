package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type MaxCntCacheDecorator struct {
	Cache  *LocalCache
	mutex  sync.RWMutex
	Cnt    int32
	MaxCnt int32
}

func NewMaxCntCache(maxCnt int32) *MaxCntCacheDecorator {
	ret := &MaxCntCacheDecorator{
		MaxCnt: maxCnt,
	}
	c := NewLocalCache(func(key string, val any) {
		atomic.AddInt32(&ret.Cnt, -1)
	})
	ret.Cache = c
	return ret
}

func (c *MaxCntCacheDecorator) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.Cache.Get(ctx, key)
	if err != nil && err != errKeyNotFound {
		return err
	}

	if err == errKeyNotFound {
		nowCnt := atomic.AddInt32(&c.Cnt, 1)

		if nowCnt > c.MaxCnt {
			c.Cnt = atomic.AddInt32(&c.Cnt, -1)
			return errors.New("cache：内存满了")
		}
	}

	return c.Cache.Set(ctx, key, val, expiration)
}

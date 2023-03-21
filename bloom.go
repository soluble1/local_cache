package cache

import (
	"context"
	"github.com/bits-and-blooms/bloom/v3"
	"time"
)

type BloomCache struct {
	bloom.BloomFilter
	Cache
	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration
}

func NewBloomCache(cache Cache, blm bloom.BloomFilter, loadFunc func(ctx context.Context, key string) (any, error)) Cache {
	return &BloomCache{
		Cache:       cache,
		BloomFilter: blm,
		LoadFunc:    loadFunc,
	}
}

func (b *BloomCache) Get(ctx context.Context, key string) (any, error) {
	val, err := b.Cache.Get(ctx, key)
	if err != nil && err != errKeyNotFound {
		return nil, err
	}
	if err == errKeyNotFound {
		exist := b.BloomFilter.Test([]byte(key))
		if exist {
			// 存在
			val, err = b.LoadFunc(ctx, key)
			if err != nil {
				return nil, err
			}
			err = b.Cache.Set(ctx, key, val, b.Expiration)
		}
	}
	return val, err
}

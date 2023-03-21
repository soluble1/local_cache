package cache

import (
	"context"
	"golang.org/x/sync/singleflight"
)

type SingleFlightCache struct {
	ReadThroughCache
}

func NewSingleFlightCache(cache Cache, loadFunc func(ctx context.Context, key string) (any, error)) Cache {
	g := &singleflight.Group{}
	return &SingleFlightCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			Loader: LoadFunc(func(ctx context.Context, key string) (any, error) {
				defer func() {
					// Forget告诉singleflight忘记一个键。未来对这个键的Do的调用将调用该函数，而不是等待先前的调用完成。
					g.Forget(key)
				}()
				// 多个 goroutine 进来这里
				// 只有一个 goroutine 会真的去执行
				val, err, _ := g.Do(key, func() (interface{}, error) {
					return loadFunc(ctx, key)
				})
				return val, err
			}),
		},
	}
}

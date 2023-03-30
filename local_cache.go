package cache

import (
	"context"
	"sync"
	"time"
)

type LocalCacheOption func(l *LocalCache)

type LocalCache struct {
	// data 存储不用sync.Map，因为在控制过期时间的时候需要一个实实在在的锁进行控制
	data  map[string]any
	mutex sync.RWMutex

	// 控制轮询的关闭
	close     chan struct{}
	closeOnce sync.Once

	// CDC 回调
	onEvicted func(key string, val any)
}

func WithOnEvicted(o func(key string, val any)) LocalCacheOption {
	return func(l *LocalCache) {
		l.onEvicted = o
	}
}

func NewLocalCache(opts ...LocalCacheOption) *LocalCache {
	l := &LocalCache{
		data:  make(map[string]any),
		close: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(l)
	}

	// 轮询检查是否有数据过期
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				cnt := 0
				l.mutex.Lock()
				for k, v := range l.data {
					itm := v.(*item)
					if itm.deadline.Before(time.Now()) {
						l.delete(k, itm.val)
					}
					cnt++
					if cnt > 1000 {
						break
					}
				}
				l.mutex.Unlock()
			case <-l.close:
				return
			}
		}
	}()

	return l
}

func (l *LocalCache) delete(key string, val any) {
	delete(l.data, key)
	if l.onEvicted != nil {
		l.onEvicted(key, val)
	}
}

// Get 懒删除：轮询 + 在 Get 的时候检查是否过期
func (l *LocalCache) Get(ctx context.Context, key string) (any, error) {
	l.mutex.RLock()
	val, ok := l.data[key]
	l.mutex.RUnlock()
	if !ok {
		return nil, errKeyNotFound
	}

	itm := val.(*item)
	// 判断过期
	if itm.deadline.Before(time.Now()) {
		l.mutex.Lock()
		defer l.mutex.Unlock()

		// double check
		val, ok = l.data[key]
		if !ok {
			return nil, errKeyNotFound
		}
		itm = val.(*item)
		if itm.deadline.Before(time.Now()) {
			l.delete(key, itm.val)
		}
		return nil, errKeyNotFound
	}
	return itm.val, nil
}

func (l *LocalCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	l.mutex.Lock()
	l.data[key] = &item{
		val:      val,
		deadline: time.Now().Add(expiration),
	}
	l.mutex.Unlock()
	return nil
}

func (l *LocalCache) Delete(ctx context.Context, key string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	val, ok := l.data[key]
	if !ok {
		return nil
	}
	l.delete(key, val.(*item).val)
	return nil
}

// Close 关闭轮询
func (l *LocalCache) Close() error {
	l.closeOnce.Do(func() {
		l.close <- struct{}{}
		close(l.close)
	})
	return nil
}

type item struct {
	val      any
	deadline time.Time
}

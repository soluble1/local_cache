package cache

import (
	"context"
	"github.com/gotomicro/ekit/list"
	"sync"
	"time"
)

type MaxMemoryCache struct {
	Cache
	mutex sync.RWMutex
	max   int64
	used  int64

	keys *list.LinkedList[string]
}

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		max:   max,
		Cache: cache,
		keys:  list.NewLinkedList[string](),
	}
	return res
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果 key 已经存在则删掉
	itm, err := m.Cache.Get(ctx, key)
	if err == nil {
		err := m.Cache.Delete(ctx, key)
		if err != nil {
			return err
		}
		m.used -= int64(len(AnyByByte(itm)))
		m.deleteKey(key)
	}

	for m.used+int64(len(AnyByByte(val))) > m.max {
		fir, err := m.keys.Get(0)
		if err != nil {
			return err
		}
		firVal, _ := m.Cache.Get(ctx, fir)
		m.keys.Delete(0)
		m.used -= int64(len(AnyByByte(firVal)))
		_ = m.Cache.Delete(ctx, fir)
	}
	err = m.Cache.Set(ctx, key, val, expiration)
	if err == nil {
		m.used += int64(len(AnyByByte(val)))
		_ = m.keys.Append(key)
	}
	return nil
}

func (m *MaxMemoryCache) Get(ctx context.Context, key string) (any, error) {
	// 加锁是为了防止遇上懒惰删除的情况，触发了删除
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, err := m.Cache.Get(ctx, key)
	if err == nil {
		// 把原本的删掉
		// 然后将 key 加到末尾
		m.deleteKey(key)
		_ = m.keys.Append(key)
	}
	return val, err
}

func (m *MaxMemoryCache) deleteKey(key string) {
	for i := 0; i < m.keys.Len(); i++ {
		ele, _ := m.keys.Get(i)
		if ele == key {
			_, _ = m.keys.Delete(i)
			return
		}
	}
}

func AnyByByte(k any) []byte {
	return []byte(k.(string))
}

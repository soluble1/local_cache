package cache

import (
	"context"
	"log"
)

type WriteBackCache struct {
	*LocalCache
}

func NewWriteBack(store func(ctx context.Context, key string, val any) error) *WriteBackCache {
	opt := WithOnEvicted(func(key string, val any) {
		err := store(context.Background(), key, val)
		if err != nil {
			log.Fatalln(err)
		}
	})

	w := &WriteBackCache{
		LocalCache: NewLocalCache(opt),
	}

	return w
}

func (w *WriteBackCache) Close() error {
	for k, v := range w.LocalCache.data {
		w.LocalCache.onEvicted(k, v.(*item).val)
	}

	return nil
}

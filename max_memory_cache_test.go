package cache

import (
	"context"
	"github.com/gotomicro/ekit/list"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMaxMemoryCache_Set(t *testing.T) {
	testCases := []struct {
		name  string
		cache func() *MaxMemoryCache

		key string
		val any

		wantKeys []string
		wantErr  error
		wantUsed int64
	}{
		{
			// 不触发淘汰
			name: "not exist",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &LocalCache{data: map[string]any{}})
				return res
			},
			key:      "key1",
			val:      "hello",
			wantKeys: []string{"key1"},
			wantUsed: 5,
		},
		{
			// 原本就有，覆盖导致 used 增加
			name: "override-incr",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &LocalCache{
					data: map[string]any{
						"key1": &item{"hello", time.Now().Add(time.Minute)},
					},
				})
				res.keys = list.NewLinkedListOf[string]([]string{"key1"})
				res.used = 5
				return res
			},
			key:      "key1",
			val:      "hello,world",
			wantKeys: []string{"key1"},
			wantUsed: 11,
		},
		{
			// 原本就有，覆盖导致 used 减少
			name: "override-decr",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &LocalCache{
					data: map[string]any{
						"key1": &item{"hello", time.Now().Add(time.Minute)},
					},
				})
				res.keys = list.NewLinkedListOf[string]([]string{"key1"})
				res.used = 5
				return res
			},
			key:      "key1",
			val:      "he",
			wantKeys: []string{"key1"},
			wantUsed: 2,
		},
		{
			// 执行淘汰，一次
			name: "delete",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(40, &LocalCache{
					data: map[string]any{
						"key1": &item{"hello, key1", time.Now().Add(time.Minute)},
						"key2": &item{"hello, key2", time.Now().Add(time.Minute)},
						"key3": &item{"hello, key3", time.Now().Add(time.Minute)},
					},
				})
				res.keys = list.NewLinkedListOf[string]([]string{"key1", "key2", "key3"})
				res.used = 33
				return res
			},
			key:      "key4",
			val:      "hello, key4",
			wantKeys: []string{"key2", "key3", "key4"},
			wantUsed: 33,
		},
		{
			// 执行淘汰，多次
			name: "delete-multi",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(40, &LocalCache{
					data: map[string]any{
						"key1": &item{"hello, key1", time.Now().Add(time.Minute)}, // 11
						"key2": &item{"hello, key2", time.Now().Add(time.Minute)},
						"key3": &item{"hello, key3", time.Now().Add(time.Minute)},
					},
				})
				res.keys = list.NewLinkedListOf[string]([]string{"key1", "key2", "key3"})
				res.used = 33
				return res
			},
			key:      "key4",
			val:      "hello, key4,hello, key4", // 23
			wantKeys: []string{"key3", "key4"},
			wantUsed: 34,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			err := cache.Set(context.Background(), tc.key, tc.val, time.Minute)
			assert.Equal(t, tc.wantKeys, cache.keys.AsSlice())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUsed, cache.used)
		})
	}
}

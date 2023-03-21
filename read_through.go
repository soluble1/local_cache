package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type ReadThroughCache struct {
	Cache
	mutex      sync.RWMutex
	Expiration time.Duration
	// 从DB中加载数据
	//LoadFunc func(ctx context.Context, key string) (any, error)
	Loader
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	// 1.读缓存
	r.mutex.RLock()
	val, err := r.Cache.Get(ctx, key)
	r.mutex.RUnlock()
	// 2.读缓存失败
	if err != nil {
		// 尽管加锁了在并发写的时候任然会导致数据不一致

		// 2.1.从 DB 中拿数据
		r.mutex.RLock()
		val, err = r.Load(ctx, key)
		r.mutex.RUnlock()

		// *不加锁* 第一个 G1 进来key1 = value1
		// 	  	   中间有人更新了数据库
		//    	   第二个 G2 进来Key1 = value2

		// 2.2.从 DB 中拿数据失败
		if err != nil {
			// LoadFunc 是数据库查询，直接返回err可能暴露数据库信息
			return nil, fmt.Errorf("cache: 无法加载数据，%w", err)
		}
		// 2.2.从 DB 中拿数据成功更新缓存
		r.mutex.Lock()
		err = r.Cache.Set(ctx, key, val, r.Expiration)
		r.mutex.Unlock()

		// *不加锁* 1.G1 先进来，数据一致
		//         2.G2 先进来，数据不一致

		if err != nil {
			// 2.2.1.刷新缓存失败记录日志
			log.Fatalln("read through cache : 刷新缓存失败！")
		}
		// 2.2.2.返回数据
		return val, nil
	}
	// 3.读缓存成功
	return val, nil
}

type Loader interface {
	Load(ctx context.Context, key string) (any, error)
	// 可扩展
}

type LoadFunc func(ctx context.Context, key string) (any, error)

func (l LoadFunc) Load(ctx context.Context, key string) (any, error) {
	return l(ctx, key)
}

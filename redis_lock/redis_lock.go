package redis_lock

import (
	"context"
	_ "embed"
	"errors"
	uuid2 "github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

/*
	Redis 使用单个 Lua 解释器去运行所有脚本，并且， Redis 也保证脚本会以原子性(atomic)的方式执行：
	当某个脚本正在运行的时候，不会有其他脚本或 Redis 命令被执行。
*/

var (
	//go:embed lua/unlock.lua
	luaUnlock string

	//go:embed lua/refresh.lua
	RefreshLock string

	//go:embed lua/lock.lua
	luaLock string

	ErrFailedToPreemptLock = errors.New("redisLock: 抢锁失败")
	// ErrLockNotHold 一般是出现在你预期你本来持有锁，结果却没有持有锁的地方
	// 比如说当你尝试释放锁的时候，可能得到这个错误
	// 这一般意味着有人绕开了 redisLock 的控制，直接操作了 Redis
	ErrLockNotHold = errors.New("redisLock: 未持有锁")
)

type Client struct {
	client redis.Cmdable
}

func NewClient(red redis.Cmdable) *Client {
	return &Client{
		client: red,
	}
}

// Lock 加锁可能遇到偶然性失败，在这种情况下尝试重试
// 1.如果失败了就马上重试
// 2.重试返回的值与上一次相同表示加锁成功，不是表示锁被人持有
func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration, retry RetryStrategy, timeout time.Duration) (*Lock, error) {
	val := uuid2.New().String()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	for {
		ctxi, cancel := context.WithTimeout(ctx, timeout)
		// 抢锁
		ok, err := c.client.Eval(ctxi, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if ok == "OK" {
			// 抢锁成功
			return &Lock{
				client:     c.client,
				key:        key,
				value:      val,
				expiration: expiration,

				over: make(chan struct{}, 1),
			}, nil
		}

		// 抢锁失败
		// 加锁超时
		if err != nil && err == context.DeadlineExceeded {
			return nil, err
		}

		// 这里表示锁被其他人持有
		t, b := retry.Next()
		if !b {
			return nil, ErrFailedToPreemptLock
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(t):

		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	// 抢锁
	val := uuid2.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		client:     c.client,
		key:        key,
		value:      val,
		expiration: expiration,

		over: make(chan struct{}, 1),
	}, nil

}

// AutoRefresh 自动刷新过期时间
func (l *Lock) AutoRefresh(internal time.Duration, timeout time.Duration) error {
	// 1.设置刷新的间隔时间
	ticker := time.NewTicker(internal)
	defer ticker.Stop()
	// 续约信号
	try := make(chan struct{}, 1)
	defer close(try)

	for {
		select {
		// 2.尝试刷新
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// 检查是否存在并且是自己的锁,是的话就刷新一下过期时间
			err := l.Refresh(ctx)
			cancel()

			if err == context.DeadlineExceeded {
				// 继续尝试刷新
				try <- struct{}{}
				continue
			}

			// 处理由于其他错误产生的失败
			if err != nil {
				return err
			}
		case <-try:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()

			if err == context.DeadlineExceeded {
				// 继续尝试刷新
				try <- struct{}{}
				continue
			}

			// 处理由于其他错误产生的失败
			if err != nil {
				return err
			}
		case <-l.over:
			// 处理退出
			return nil
		}
	}
}

// Refresh 刷新一次过期时间
func (l *Lock) Refresh(ctx context.Context) error {
	// 1.检查是否存在并且是自己的锁,是的话就刷新一下过期时间
	ok, err := l.client.Eval(ctx, RefreshLock, []string{l.key}, l.value, l.expiration.Seconds()).Uint64()
	if err != nil {
		return err
	}

	if ok != 1 {
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) UnLock(ctx context.Context) error {
	// 解锁的时候将续约关闭
	l.onceOver.Do(func() {
		close(l.over)
	})

	// 根据 value 值来检查这把锁是不是自身的锁
	ret, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Uint64()
	if err != nil {
		return err
	}

	if ret != 1 {
		return ErrLockNotHold
	}

	return nil
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration

	over     chan struct{}
	onceOver sync.Once
}

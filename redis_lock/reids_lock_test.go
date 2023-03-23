package redis_lock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"testing"
	"time"
)

func TestRedisLockTryLock(t *testing.T) {
	c := &Client{
		client: redis.NewClient(&redis.Options{
			Addr:     ":6379",
			Password: "",
			DB:       0,
		}),
	}

	lock, err := c.TryLock(context.Background(), "mykey", time.Minute)
	if err != nil {
		log.Println(err)
	}

	log.Printf("key: %s   val: %s\n", lock.key, lock.value)

	err = lock.UnLock(context.Background())
	if err != nil {
		log.Println(err)
	}
	log.Printf("key: %s   val: %s\n", lock.key, lock.value)

}

func Run(namea string, key string, rlock *Client) {
	name := namea
	for {
		retry := &FixIntervalRetry{
			// 每0.1秒尝试获取锁
			Interval: time.Second / 10,
			Max:      10,
		}
		ctx := context.Background()
		lock, err := rlock.Lock(ctx, key, time.Second*2, retry, time.Second/10)
		if err != nil {
			log.Printf("抢锁失败拉************%s     %s", name, err)
			continue
		}

		go func() {
			// 开启自动续约   续约间隔时间 接受在多少时间内续约成功
			err = lock.AutoRefresh(time.Second*1, time.Second)
			if err != nil {
				log.Println(err)
				return
			}
		}()

		// 执行代码
		log.Printf("%s\t抢锁成功并开启自动续约-----------------------", name)
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			log.Printf("%s\t执行了第 %d 秒  l.value: %s", name, i, lock.value)
		}
		log.Printf("%s\t执行完了-----------------------------------", name)
		err = lock.UnLock(ctx)
		if err != nil {
			log.Println(err)
		}
	}
}

func TestRedisLock(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "120.46.196.48:6379",
		Password: "",
		DB:       0,
	})
	rlock := NewClient(client)

	key := "mykey"

	go func() {
		Run("xiaolong", key, rlock)
	}()

	go func() {
		Run("laozhang", key, rlock)
	}()

	go func() {
		Run("majundashi", key, rlock)
	}()

	time.Sleep(time.Minute)
}

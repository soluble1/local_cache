package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	r := redis.NewClient(&redis.Options{
		Addr:     ":6379",
		Password: "",
		DB:       0,
	})

	local := NewRedisCache(r)

	local.Set(context.Background(), "xiao", "xiao long ren", time.Second*3)
	local.Set(context.Background(), "ma", "ma jun da shi", time.Second*6)
	local.Set(context.Background(), "lao", "lao zhang", time.Second*2)
	local.Set(context.Background(), "test", "this is test", time.Minute*2)

	g := sync.WaitGroup{}
	g.Add(3)

	go func() {
		for {
			time.Sleep(time.Second)
			xiao, err := local.Get(context.Background(), "xiao")
			if err != nil {
				log.Println(err)
				g.Done()
				return
			}
			fmt.Println(xiao)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			ma, err := local.Get(context.Background(), "ma")
			if err != nil {
				log.Println(err)
				g.Done()
				return
			}
			fmt.Println(ma)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			lao, err := local.Get(context.Background(), "lao")
			if err != nil {
				log.Println(err)
				g.Done()
				return
			}
			fmt.Println(lao)
		}
	}()
	g.Wait()

}

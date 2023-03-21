package cache

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
	"testing"
	"time"
)

func TestReadThroughCache(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(:3307)/mydb?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	onEvicted := WithOnEvicted(func(key string, val any) {
		log.Printf("这个现在被删除拉 : %s", key)
	})

	rtc := &ReadThroughCache{
		Cache: NewLocalCache(onEvicted),
		Loader: LoadFunc(func(ctx context.Context, key string) (any, error) {
			row := db.QueryRow("select `password` from `user` where `user_name`=?", key)
			var val any
			err := row.Scan(&val)
			return val, err
		}),
	}

	//rtc.Cache.Set(context.Background(), "xiao", "xiao long ren", time.Second*3)
	//rtc.Cache.Set(context.Background(), "ma", "ma jun da shi", time.Second*6)
	//rtc.Cache.Set(context.Background(), "lao", "lao zhang", time.Second*2)

	g := sync.WaitGroup{}
	g.Add(3)

	go func() {
		for {
			time.Sleep(time.Second)
			xiao, err := rtc.Get(context.Background(), "xiao")
			if err != nil {
				//log.Println(err)
				g.Done()
				return
			}
			res := string(xiao.([]uint8))
			fmt.Println(res)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			ma, err := rtc.Get(context.Background(), "ma")
			if err != nil {
				//log.Println(err)
				g.Done()
				return
			}
			res := string(ma.([]uint8))
			fmt.Println(res)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			lao, err := rtc.Get(context.Background(), "lao")
			if err != nil {
				//log.Println(err)
				g.Done()
				return
			}
			res := string(lao.([]uint8))
			fmt.Println(res)
		}
	}()
	g.Wait()

}

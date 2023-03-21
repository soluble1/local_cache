package cache

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestSingleFlightCache(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(120.46.196.48:3307)/mydb?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	onEvicted := WithOnEvicted(func(key string, val any) {
		log.Printf("这个现在被删除拉 : %s", key)
	})

	local := NewLocalCache(onEvicted)

	sf := NewSingleFlightCache(local, func(ctx context.Context, key string) (any, error) {
		//log.Println("loadFunc")
		//time.Sleep(time.Second * 3)
		log.Println("----------抢了一次锁-------------")
		row := db.QueryRow("select `password` from `user` where `user_name`=?", key)
		var val any
		err := row.Scan(&val)
		return val, err
	})

	keys := "this is con not exist key"

	for i := 0; i < 8; i++ {
		go func() {
			for {
				time.Sleep(time.Second / 10)
				_, err := sf.Get(context.Background(), keys)
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}

	time.Sleep(time.Second * 10)
}

// 不使用 singleFlightCache
func TestSingleFlightCacheV1(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(120.46.196.48:3307)/mydb?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	onEvicted := WithOnEvicted(func(key string, val any) {
		log.Printf("这个现在被删除拉 : %s", key)
	})

	local := NewLocalCache(onEvicted)

	rtc := &ReadThroughCache{
		Cache: local,
		Loader: LoadFunc(func(ctx context.Context, key string) (any, error) {
			//log.Println("loadFunc")
			//time.Sleep(time.Second * 3)
			row := db.QueryRow("select `password` from `user` where `user_name`=?", key)
			var val any
			err := row.Scan(&val)
			return val, err
		}),
	}

	keys := "this is con not exist key"

	cnt := 0
	for i := 0; i < 8; i++ {
		go func() {
			for {
				time.Sleep(time.Second / 10)
				_, err2 := rtc.Get(context.Background(), keys)
				if err2 != nil {
					fmt.Println(cnt)
					cnt++
					log.Println(err2)
				}
			}
		}()
	}

	time.Sleep(time.Second * 10)
}

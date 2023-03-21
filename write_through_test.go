package cache

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"log"
	"testing"
	"time"
)

func TestWriteThroughCache(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(:3307)/mydb?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	onEvicted := WithOnEvicted(func(key string, val any) {
		log.Printf("这个现在被删除拉 : %s", key)
	})

	wtc := &WriteThroughCache{
		Cache: NewLocalCache(onEvicted),
		StoreFunc: func(ctx context.Context, key string, val any) error {
			_, err := db.Exec("INSERT INTO `user`(`user_name`,`password`) values(?,?)", key, val)
			if err != nil {
				return err
			}
			log.Printf("这个key：%s   插入数据库成功拉\n", key)
			return nil
		},
	}

	go func() {
		for {
			time.Sleep(time.Second)
			uuid := uuid.New()
			key := uuid.String()
			err := wtc.Set(context.Background(), key, "this is uuid", time.Second/2)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	time.Sleep(time.Minute * 5)
}

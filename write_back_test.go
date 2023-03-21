package cache

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"log"
	"testing"
	"time"
)

func TestWriteBackCache(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(:3307)/mydb?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	wbc := NewWriteBack(func(ctx context.Context, key string, val any) error {
		_, err := db.Exec("INSERT INTO `user`(`user_name`,`password`) values(?,?)", key, val)
		if err != nil {
			return err
		}
		log.Printf("这个key过期插入数据库成功拉：%s\n", key)
		return nil
	})
	defer func() {
		log.Println("----------关闭数据库咯------------")
		err := wbc.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			uuid := uuid.New()
			key := uuid.String()
			err := wbc.Set(context.Background(), key, "this is uuid", time.Second*2)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	time.Sleep(time.Second * 10)
}

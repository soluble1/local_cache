package cache

import (
	"context"
	uuid2 "github.com/google/uuid"
	"log"
	"testing"
	"time"
)

func TestMaxCntCacheDecorator(t *testing.T) {
	mccd := NewMaxCntCache(5)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		log.Println(i)
		uuid, _ := uuid2.NewUUID()
		err := mccd.Set(context.Background(), uuid.String(), "this is test", time.Minute)
		if err != nil {
			log.Println(err)
			break
		}
	}

	time.Sleep(time.Second * 20)
}

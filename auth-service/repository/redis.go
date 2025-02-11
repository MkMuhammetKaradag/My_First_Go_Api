package repository

import (
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis"
	// "github.com/redis/go-redis"
)

var RedisClient *redis.Client
var once sync.Once

func ConnectRedis(addr string, db int) {
	once.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr: addr, // "localhost:6379",
			DB:   db,
		})

		_, err := RedisClient.Ping().Result()
		if err != nil {
			log.Fatalf("Redis bağlantı hatası: %v", err)
		}

		fmt.Println("Redis bağlantısı başarılı")
	})
}

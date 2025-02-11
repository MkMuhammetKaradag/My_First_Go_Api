package database

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis"
)

var (
	RedisClient      *redis.Client
	onceRedisConnect sync.Once
)


func ConnectRedis(addr string, db int) {
	onceRedisConnect.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr: addr, // "localhost:6379",
			// Password: os.Getenv("REDIS_PASSWORD"), 
			DB: db, // 0,
		})

		_, err := RedisClient.Ping().Result()
		if err != nil {
			panic(fmt.Sprintf("Redis bağlantı hatası: %v", err))
		}

		fmt.Println("Redis bağlantısı başarılı")
	})
}

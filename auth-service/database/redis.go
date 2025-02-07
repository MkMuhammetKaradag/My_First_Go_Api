package database

import (
	"fmt"

	"github.com/go-redis/redis"
	// "github.com/redis/go-redis"
)

var RedisClient *redis.Client

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Bağlantıyı test et
	_, err := RedisClient.Ping().Result() // Buradaki `Ping()` çağrısını değiştirdik
	if err != nil {
		fmt.Println("Redis bağlantı hatası:", err)
		return
	}

	fmt.Println("Redis bağlantısı başarılı")
}

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

// Redis'e bağlan
func ConnectRedis(addr string, db int) error {
	var err error

	onceRedisConnect.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr: addr, // Örn: "localhost:6379"
			DB:   db,   // 0 numaralı database
		})

		_, connErr := RedisClient.Ping().Result()
		if connErr != nil {
			err = fmt.Errorf("Redis bağlantı hatası: %w", connErr)
		} else {
			fmt.Println("Redis bağlantısı başarılı")
		}
	})

	return err
}

// Redis bağlantısını kapatma
func DisconnectRedis() error {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			return fmt.Errorf("Redis bağlantısı kapatılamadı: %w", err)
		}
		fmt.Println("Redis bağlantısı kapatıldı.")
	}
	return nil
}

// Redis bağlantısının aktif olup olmadığını kontrol et
func IsRedisConnected() bool {
	if RedisClient == nil {
		return false
	}
	_, err := RedisClient.Ping().Result()
	return err == nil
}

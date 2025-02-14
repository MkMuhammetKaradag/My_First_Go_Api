// repository/redis_repository.go
package redisrepo

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type RedisRepository struct {
	Client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{Client: client}
}

// Session işlemleri
func (r *RedisRepository) SetSession(key string, userData map[string]string, expiration time.Duration) error {
	jsonData, err := json.Marshal(userData)
	if err != nil {
		return err
	}
	return r.Client.Set(key, jsonData, expiration).Err()
}

func (r *RedisRepository) GetSession(key string) (map[string]string, error) {
	data, err := r.Client.Get(key).Result()
	if err != nil {
		return nil, err
	}

	var userData map[string]string
	if err := json.Unmarshal([]byte(data), &userData); err != nil {
		return nil, err
	}
	return userData, nil
}

func (r *RedisRepository) DeleteSession(key string) error {
	return r.Client.Del(key).Err()
}

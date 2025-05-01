package storage

import "github.com/redis/go-redis/v9"

type CacheStorage struct {
	client *redis.Client
}

func NewCacheStorage(password string) CacheStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: password,
		DB:       0,
	})

	return CacheStorage{client: rdb}
}

func (c *CacheStorage) Close() {
	if c != nil {
		c.client.Close()
	}
}

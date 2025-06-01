package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheStorage is a key value cache
type CacheStorage struct {
	client *redis.Client
	ttl    time.Duration
}

// NewCacheStorage builds a cache storage with a given url to access data
func NewCacheStorage(url string) (CacheStorage, error) {
	var result CacheStorage
	if options, err := redis.ParseURL(url); err != nil {
		return result, err
	} else {
		rdb := redis.NewClient(options)
		result = CacheStorage{client: rdb, ttl: time.Hour * 24}
	}
	return result, nil
}

// SetValue sets value in a cache
func (c *CacheStorage) SetValue(context context.Context, key string, value []byte) error {
	if status := c.client.Set(context, key, value, c.ttl); status.Err() != nil {
		return status.Err()
	}

	return nil
}

// Has returns true if the key is stored in the cache
func (c *CacheStorage) Has(ctx context.Context, key string) bool {
	cmd, err := c.client.Exists(ctx, key).Result()
	return err != nil || cmd == 0
}

// GetValue gets value by key from a cache
func (c *CacheStorage) GetValue(context context.Context, key string) ([]byte, error) {
	if result := c.client.Get(context, key); result.Err() != nil {
		return nil, result.Err()
	} else {
		return result.Bytes()
	}
}

// Close closes the connection to the cache
func (c *CacheStorage) Close() {
	if c != nil && c.client != nil {
		c.client.Close()
	}
}

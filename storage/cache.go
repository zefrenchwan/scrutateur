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
func (c *CacheStorage) SetValue(context context.Context, key, value string) error {
	if status := c.client.Set(context, key, value, c.ttl); status.Err() != nil {
		return status.Err()
	}

	return nil
}

// GetValue gets value by key from a cache
func (c *CacheStorage) GetValue(context context.Context, key string) (string, error) {
	if result := c.client.Get(context, key); result.Err() != nil {
		return "", result.Err()
	} else {
		return result.String(), nil
	}
}

// Close closes the connection to the cache
func (c *CacheStorage) Close() {
	if c != nil {
		c.client.Close()
	}
}

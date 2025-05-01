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

// NewCacheStorage builds a cache storage with a given password to access data
func NewCacheStorage(password string) CacheStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: password,
		DB:       0,
	})

	return CacheStorage{client: rdb, ttl: time.Hour * 24}
}

// SetSessionForUser sets the new value for a session id (defined at user level)
func (c *CacheStorage) SetSessionForUser(context context.Context, sessionId string, session []byte) error {
	if status := c.client.Set(context, sessionId, session, c.ttl); status.Err() != nil {
		return status.Err()
	}

	return nil
}

// GetSessionForUser loads session data based on its session id
func (c *CacheStorage) GetSessionForUser(context context.Context, sessionId string) ([]byte, error) {
	if result := c.client.Get(context, sessionId); result.Err() != nil {
		return nil, result.Err()
	} else {
		return result.Bytes()
	}
}

// Close closes the connection to the cache
func (c *CacheStorage) Close() {
	if c != nil {
		c.client.Close()
	}
}

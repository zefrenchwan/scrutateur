package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zefrenchwan/scrutateur.git/dto"
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

// SetSessionForUser sets the new value for a session id (defined at user level)
func (c *CacheStorage) SetSessionForUser(context context.Context, sessionId string, session dto.Session) error {
	if value, err := session.Serialize(); err != nil {
		return err
	} else if status := c.client.Set(context, sessionId, value, c.ttl); status.Err() != nil {
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

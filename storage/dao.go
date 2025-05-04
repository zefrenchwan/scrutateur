package storage

import (
	"context"
)

// Dao decorates any storage system (relational, cache, etc)
type Dao struct {
	rdb   DbStorage
	cache CacheStorage
}

// NewDao returns a new dao for that db user and password
func NewDao(user, password string, cachePassword string) (Dao, error) {
	var dao Dao
	if db, err := NewDbStorage(user, password); err != nil {
		return dao, err
	} else {
		cacheClient := NewCacheStorage(cachePassword)
		return Dao{rdb: db, cache: cacheClient}, nil
	}
}

// Close all the storage systems
func (d *Dao) Close() {
	if d != nil {
		d.rdb.Close()
		d.cache.Close()
	}
}

// ValidateUser returns true if login and password are a valid user auth info.
func (d *Dao) ValidateUser(ctx context.Context, login string, password string) (bool, error) {
	return d.rdb.ValidateUser(ctx, login, password)
}

// SetSessionForUser registers a session
func (d *Dao) SetSessionForUser(context context.Context, sessionId string, session []byte) error {
	return d.cache.SetSessionForUser(context, sessionId, session)
}

// GetSessionForUser gets a session by id
func (d *Dao) GetSessionForUser(context context.Context, sessionId string) ([]byte, error) {
	return d.cache.GetSessionForUser(context, sessionId)
}

// GetUserRoles loads roles for a user.
// Result is a map of role and score
func (d *Dao) GetUserRoles(context context.Context, user string) (map[string]int, error) {
	return d.rdb.LoadUserRoles(context, user)
}

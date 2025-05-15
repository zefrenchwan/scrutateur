package storage

import (
	"context"

	"github.com/zefrenchwan/scrutateur.git/dto"
)

type DaoOptions struct {
	RedisURL      string
	PostgresqlURL string
}

// Dao decorates any storage system (relational, cache, etc)
type Dao struct {
	rdb   DbStorage
	cache CacheStorage
}

// NewDao returns a new dao for those connection parameters
func NewDao(options DaoOptions) (Dao, error) {
	var dao Dao
	if db, err := NewDbStorage(options.PostgresqlURL); err != nil {
		return dao, err
	} else if cacheClient, err := NewCacheStorage(options.RedisURL); err != nil {
		return dao, err
	} else {
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
func (d *Dao) SetSessionForUser(context context.Context, sessionId string, session dto.Session) error {
	return d.cache.SetSessionForUser(context, sessionId, session)
}

// GetSessionForUser gets a session by id
func (d *Dao) GetSessionForUser(context context.Context, sessionId string) (dto.Session, error) {
	if raw, err := d.cache.GetSessionForUser(context, sessionId); err != nil {
		return dto.Session{}, err
	} else {
		return dto.SessionLoad(raw)
	}
}

// GetUserGrantedAccess returns, for a user, all the rules conditions to access a resource
func (d *Dao) GetUserGrantedAccess(context context.Context, user string) ([]dto.GrantAccessForResource, error) {
	return d.rdb.GetUserGrantedAccess(context, user)
}

// UpsertUser creates an user in database with that password if it does not exist, or changes current password
func (d *Dao) UpsertUser(ctx context.Context, username, password string) error {
	return d.rdb.UpsertUser(ctx, username, password)
}

// GrantAccessToGroupOfResources sets roles for user to that group of resources
func (d *Dao) GrantAccessToGroupOfResources(ctx context.Context, username string, roles []dto.GrantRole, group string) error {
	return d.rdb.GrantAccessToGroupOfResources(ctx, username, roles, group)
}

// RemoveAccessToGroupOfResources removes access rights for that user to a given group of resources
func (d *Dao) RemoveAccessToGroupOfResources(ctx context.Context, username string, group string) error {
	return d.rdb.RemoveAccessToGroupOfResources(ctx, username, group)
}

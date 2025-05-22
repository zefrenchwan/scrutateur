package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/zefrenchwan/scrutateur.git/dto"
)

type DaoOptions struct {
	RedisURL      string
	PostgresqlURL string
}

// Dao decorates any storage system (relational, cache, etc)
type Dao struct {
	rdb    DbStorage
	cache  CacheStorage
	logger *log.Logger
}

// NewDao returns a new dao for those connection parameters
func NewDao(options DaoOptions, logger *log.Logger) (Dao, error) {
	var dao Dao
	if len(options.PostgresqlURL) == 0 {
		return dao, fmt.Errorf("missing postgres configuration")
	} else if db, err := NewDbStorage(options.PostgresqlURL); err != nil {
		return dao, err
	} else if len(options.RedisURL) == 0 {
		logger.Println("REDIS CACHE INACTIVE. Dao will only manage relational database")
		return Dao{rdb: db, logger: logger}, nil
	} else if cacheClient, err := NewCacheStorage(options.RedisURL); err != nil {
		return dao, err
	} else {
		return Dao{rdb: db, cache: cacheClient, logger: logger}, nil
	}
}

// Close all the storage systems
func (d *Dao) Close() {
	if d != nil {
		d.rdb.Close()
		d.cache.Close()
	}
}

// Log logs info for dao
func (d *Dao) Log(messages ...any) {
	d.logger.Println(messages...)
}

// ValidateUser returns true if login and password are a valid user auth info.
func (d *Dao) ValidateUser(ctx context.Context, login string, password string) (bool, error) {
	return d.rdb.ValidateUser(ctx, login, password)
}

// SetValue sets value in a cache
func (d *Dao) SetValue(context context.Context, key, value string) error {
	return d.cache.SetValue(context, key, value)
}

// GetValue gets value by key in a cache
func (d *Dao) GetValue(context context.Context, key string) (string, error) {
	return d.cache.GetValue(context, key)
}

// GetUserGrantedAccess returns, for a user, all the rules conditions to access a resource
func (d *Dao) GetUserGrantedAccess(context context.Context, user string) ([]dto.GrantAccessForResource, error) {
	return d.rdb.GetUserGrantedAccess(context, user)
}

// GetUserGrantAccessPerGroup returns, for each resources group, all roles for that group that the user was granted
func (d *Dao) GetUserGrantAccessPerGroup(ctx context.Context, username string) (map[string][]dto.GrantRole, error) {
	return d.rdb.GetUserGrantAccessPerGroup(ctx, username)
}

// UpsertUser creates an user in database with that password if it does not exist, or changes current password
func (d *Dao) UpsertUser(ctx context.Context, username, password string) error {
	return d.rdb.UpsertUser(ctx, username, password)
}

// DeleteUser deletes user regardless user's access rights
func (d *Dao) DeleteUser(ctx context.Context, username string) error {
	return d.rdb.DeleteUser(ctx, username)
}

// GrantAccessToGroupOfResources sets roles for user to that group of resources
func (d *Dao) GrantAccessToGroupOfResources(ctx context.Context, username string, roles []dto.GrantRole, group string) error {
	if len(roles) == 0 {
		return fmt.Errorf("to grant access, one role at least is necessary")
	}

	return d.rdb.GrantAccessToGroupOfResources(ctx, username, roles, group)
}

// RemoveAccessToGroupOfResources removes access rights for that user to a given group of resources
func (d *Dao) RemoveAccessToGroupOfResources(ctx context.Context, username string, group string) error {
	return d.rdb.RemoveAccessToGroupOfResources(ctx, username, group)
}

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

// LogEvent logs an event on the events schema
func (d *Dao) LogEvent(ctx context.Context, login, actionType, actionDescription string, parameters []string) error {
	return d.rdb.LogEvent(ctx, login, actionType, actionDescription, parameters)
}

// CreateUsersGroup creates a group of users.
// Login is the user that created the group, and that user has access rights to set
func (d *Dao) CreateUsersGroup(ctx context.Context, login, groupName string, roles []dto.GrantRole) error {
	return d.rdb.CreateUsersGroup(ctx, login, groupName, roles)
}

// ListUserGroupsForSpecificUser returns the groups an user is in
func (d *Dao) ListUserGroupsForSpecificUser(ctx context.Context, login string) (map[string][]dto.GrantRole, error) {
	return d.rdb.ListUserGroupsForSpecificUser(ctx, login)
}

// GetGroupAuthForUser returns, for a specific group and user, user's auth (if any)
func (d *Dao) GetGroupAuthForUser(ctx context.Context, login, group string) ([]dto.GrantRole, error) {
	return d.rdb.GetGroupAuthForUser(ctx, login, group)
}

// SetGroupAuthForUser sets auth within a group for a given user, granted by a creator
func (d *Dao) SetGroupAuthForUser(ctx context.Context, creator, user, group string, roles []dto.GrantRole) error {
	return d.rdb.SetGroupAuthForUser(ctx, creator, user, group, roles)
}

// RevokeUserInGroup removes an user in a group
func (d *Dao) RevokeUserInGroup(ctx context.Context, user, group string) error {
	return d.rdb.RevokeUserInGroup(ctx, user, group)
}

// DeleteUsersGroup just deletes a group of users
func (d *Dao) DeleteUsersGroup(ctx context.Context, name string) error {
	return d.rdb.DeleteUsersGroup(ctx, name)
}

// ValidateUser returns true if login and password are a valid user auth info.
func (d *Dao) ValidateUser(ctx context.Context, login string, password string) (bool, error) {
	if resp, err := d.rdb.ValidateUser(ctx, login, password); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return false, err
	} else {
		return resp, err
	}
}

// SetValue sets value in a cache
func (d *Dao) SetValue(context context.Context, key, value string) error {
	return d.cache.SetValue(context, key, []byte(value))
}

// GetValue gets value by key in a cache
func (d *Dao) GetValue(context context.Context, key string) (string, error) {
	if rawContent, err := d.cache.GetValue(context, key); err != nil {
		return "", err
	} else {
		return string(rawContent), nil
	}
}

// GetFeaturesSet returns all the resources group names (ordered by name)
func (d *Dao) GetFeaturesSet(ctx context.Context) ([]string, error) {
	if resp, err := d.rdb.GetFeaturesSet(ctx); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return nil, err
	} else {
		return resp, err
	}
}

// GetUserGrantedAccess returns, for a user, all the rules conditions to access a resource
func (d *Dao) GetUserGrantedAccess(context context.Context, user string) ([]dto.GrantAccessForResource, error) {
	if resp, err := d.rdb.GetUserGrantedAccess(context, user); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return nil, err
	} else {
		return resp, err
	}
}

// GetUserRolesPerFeature returns, for each resources group, all roles for that group that the user was granted
func (d *Dao) GetUserRolesPerFeature(ctx context.Context, username string) (map[string][]dto.GrantRole, error) {
	if resp, err := d.rdb.GetUserRolesPerFeature(ctx, username); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return nil, err
	} else {
		return resp, err
	}
}

// UpsertUser creates an user in database with that password if it does not exist, or changes current password
func (d *Dao) UpsertUser(ctx context.Context, username, password string) error {
	if err := d.rdb.UpsertUser(ctx, username, password); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return err
	} else {
		return err
	}
}

// DeleteUser deletes user regardless user's access rights
func (d *Dao) DeleteUser(ctx context.Context, username string) error {
	if err := d.rdb.DeleteUser(ctx, username); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return err
	} else {
		return err
	}
}

// GrantAccessToFeatures sets access on groups for a given user.
// The access parameter is a map of groups (should exist) and values are the roles to set.
// Note that roles are the only roles set (no append)
func (d *Dao) GrantAccessToFeatures(ctx context.Context, username string, access map[string][]dto.GrantRole) error {
	if err := d.rdb.GrantAccessToFeatures(ctx, username, access); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return err
	} else {
		return err
	}
}

// RemoveAccessToFeature removes access rights for that user to a given group of resources
func (d *Dao) RemoveAccessToFeature(ctx context.Context, username string, group string) error {
	if err := d.rdb.RemoveAccessToFeature(ctx, username, group); err != nil {
		d.logger.Println("DAO: ERROR", err)
		return err
	} else {
		return err
	}
}

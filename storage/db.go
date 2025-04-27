package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Dao decorates GORM
type Dao struct {
	db *gorm.DB
}

// NewDao returns a new dao for that db user and password
func NewDao(user, password string) (Dao, error) {
	var dao Dao
	// connect using gorm
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=capi port=5432", user, password)
	if db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		return dao, err
	} else {
		dao = Dao{db}
	}

	return dao, nil
}

// ValidateUser returns true if login and password are a valid user auth info.
// Note that we don't manipulate directly password, but its hashed version
func (d Dao) ValidateUser(login, hashedPassword string) (bool, error) {
	// Open bar for now, will change
	return true, nil
}

package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Dao struct {
	db *gorm.DB
}

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

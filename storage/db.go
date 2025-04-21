package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CreateDatabase(user, password string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=capi port=5432", user, password)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Dao decorates GORM
type Dao struct {
	db *pgxpool.Pool
}

// NewDao returns a new dao for that db user and password
func NewDao(user, password string) (Dao, error) {
	var dao Dao
	// connect using gorm
	configuration := DatabaseConfiguration(user, password)
	if db, err := CreateConnectionsPool(configuration); err != nil {
		return dao, err
	} else {
		dao = Dao{db}
	}

	return dao, nil
}

// Close the underlying pool
func (d *Dao) Close() {
	if d != nil && d.db != nil {
		d.db.Close()
	}
}

// ValidateUser returns true if login and password are a valid user auth info.
func (d *Dao) ValidateUser(login string, password string) (bool, error) {
	var result bool
	row := d.db.QueryRow(context.Background(), "select * from auth.validate_auth($1,$2)", login, password)
	if err := row.Scan(&result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

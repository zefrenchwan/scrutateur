package storage

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DbStorage decorates pgx
type DbStorage struct {
	db *pgxpool.Pool
}

// NewDbStorage starts a connection pool to the relational storage
func NewDbStorage(user, password string) (DbStorage, error) {
	var storage DbStorage
	// connect using gorm
	configuration := DatabaseConfiguration(user, password)
	if db, err := pgxpool.NewWithConfig(context.Background(), configuration); err != nil {
		return storage, err
	} else {
		storage = DbStorage{db}
	}

	return storage, nil
}

// Close the underlying pool
func (d *DbStorage) Close() {
	if d != nil && d.db != nil {
		d.db.Close()
	}
}

// Thanks to
// https://medium.com/@neelkanthsingh.jr/understanding-database-connection-pools-and-the-pgx-library-in-go-3087f3c5a0c
// That details the code and explanation

func DatabaseConfiguration(user, password string) *pgxpool.Config {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	// Your own Database URL
	var DATABASE_URL string = "postgres://" + user + ":" + password + "@localhost:5432/capi"

	dbConfig, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		log.Fatal("Failed to create a config, error: ", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	return dbConfig
}

// ValidateUser returns true if login and password are a valid user auth info.
func (d *DbStorage) ValidateUser(ctx context.Context, login string, password string) (bool, error) {
	var result bool
	row := d.db.QueryRow(ctx, "select * from auth.validate_auth($1,$2)", login, password)
	if err := row.Scan(&result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

// LoadUserRoles gets all the roles of a user
func (d DbStorage) LoadUserRoles(ctx context.Context, user string) (map[string]int, error) {
	result := make(map[string]int)
	if rows, err := d.db.Query(ctx, "select role_name, role_score from auth.get_roles_for_user($1)", user); err != nil {
		return result, err
	} else if rows == nil {
		return result, nil
	} else {
		defer rows.Close()

		for rows.Next() {
			if rows.Err() != nil {
				return result, err
			}

			var role_name string
			var role_score int
			rows.Scan(&role_name, &role_score)
			result[role_name] = role_score
		}
	}

	return result, nil
}

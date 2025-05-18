package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zefrenchwan/scrutateur.git/dto"
)

// DbStorage decorates pgx
type DbStorage struct {
	db *pgxpool.Pool
}

// NewDbStorage starts a connection pool to the relational storage
func NewDbStorage(url string) (DbStorage, error) {
	var storage DbStorage
	// connect using gorm
	configuration := DatabaseConfiguration(url)
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

func DatabaseConfiguration(url string) *pgxpool.Config {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	dbConfig, err := pgxpool.ParseConfig(url)
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

// GetUserGrantedAccess gets all the grants access for a user
func (d DbStorage) GetUserGrantedAccess(ctx context.Context, user string) ([]dto.GrantAccessForResource, error) {
	var result []dto.GrantAccessForResource
	if rows, err := d.db.Query(ctx, "select operator, template_url, roles from auth.get_grants_for_user($1) ", user); err != nil {
		return result, err
	} else if rows == nil {
		return result, nil
	} else {
		defer rows.Close()

		for rows.Next() {
			if rows.Err() != nil {
				return result, err
			}

			var operator string
			var template_url string
			roles := []string{}
			if err := rows.Scan(&operator, &template_url, &roles); err != nil {
				return result, err
			} else if parsedRoles, err := dto.ParseGrantRoles(roles); err != nil {
				return result, err
			} else if op, err := dto.ParseGrantOperator(operator); err != nil {
				return result, err
			} else {
				result = append(result, dto.GrantAccessForResource{Operator: op, Template: template_url, UserRoles: parsedRoles})
			}
		}
	}

	return result, nil
}

// UpsertUser creates an user in database with that password if it does not exist, or changes current password
func (d DbStorage) UpsertUser(ctx context.Context, username, password string) error {
	if _, err := d.db.Exec(ctx, "call auth.upsert_user_auth($1, $2)", username, password); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes user regardless user's access rights
func (d DbStorage) DeleteUser(ctx context.Context, username string) error {
	_, err := d.db.Exec(ctx, "call auth.delete_user($1)", username)
	return err
}

// GrantAccessToGroupOfResources sets roles for user to that group of resources
func (d DbStorage) GrantAccessToGroupOfResources(ctx context.Context, username string, roles []dto.GrantRole, group string) error {
	if len(roles) == 0 {
		return fmt.Errorf("no role for that grant request")
	}

	mapping := make([]string, len(roles))
	for index, value := range roles {
		mapping[index] = string(value)
	}

	_, err := d.db.Exec(ctx, "call  auth.grant_group_access_to_user($1,$2,$3)", username, mapping, group)
	return err
}

// RemoveAccessToGroupOfResources removes access rights for that user to a given group of resources
func (d DbStorage) RemoveAccessToGroupOfResources(ctx context.Context, username string, group string) error {
	_, err := d.db.Exec(ctx, "call  auth.remove_access_from_group_to_user($1,$2)", username, group)
	return err
}

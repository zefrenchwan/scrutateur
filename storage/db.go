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

// LogEvent logs an event on the events schema
func (d *DbStorage) LogEvent(ctx context.Context, login, actionType, actionDescription string, parameters []string) error {
	_, err := d.db.Exec(ctx, "call evt.log_action($1, $2, $3, $4)", login, actionType, actionDescription, parameters)
	return err
}

// CreateUsersGroup creates a group of users, from that login, with initial auth
func (d *DbStorage) CreateUsersGroup(ctx context.Context, login, name string, roles []dto.GrantRole) error {
	_, err := d.db.Exec(ctx, "call orgs.add_group($1,$2,$3)", login, name, roles)
	return err
}

// GetGroupAuthForUser returns, for a specific group and user, user's auth (if any)
func (d *DbStorage) GetGroupAuthForUser(ctx context.Context, login, group string) ([]dto.GrantRole, error) {
	var result []dto.GrantRole
	if rows, err := d.db.Query(ctx, "select local_roles from orgs.get_groups_for_user($1) where group_name = $2", login, group); err != nil {
		return result, err
	} else if rows == nil {
		return result, nil
	} else {
		defer rows.Close()

		for rows.Next() {
			if rows.Err() != nil {
				return result, err
			}

			var roles []string
			if err := rows.Scan(&roles); err != nil {
				return result, err
			} else if len(roles) == 0 {
				return result, nil
			} else if res, err := dto.ParseGrantRoles(roles); err != nil {
				return result, nil
			} else {
				result = res
				return result, nil
			}
		}

		return result, nil
	}
}

// ListUserGroupsForSpecificUser returns the groups an user is in
func (d *DbStorage) ListUserGroupsForSpecificUser(ctx context.Context, login string) (map[string][]dto.GrantRole, error) {
	var result map[string][]dto.GrantRole
	if rows, err := d.db.Query(ctx, "select group_name, local_roles from orgs.get_groups_for_user($1) ", login); err != nil {
		return result, err
	} else if rows == nil {
		return result, nil
	} else {
		defer rows.Close()

		result := make(map[string][]dto.GrantRole)
		for rows.Next() {
			if rows.Err() != nil {
				return result, err
			}

			var name string
			var values []string
			if err := rows.Scan(&name, &values); err != nil {
				return result, err
			} else if roles, err := dto.ParseGrantRoles(values); err != nil {
				return result, err
			} else {
				result[name] = roles
			}
		}

		return result, nil
	}
}

// DeleteUsersGroup just deletes a group of users
func (d *DbStorage) DeleteUsersGroup(ctx context.Context, name string) error {
	_, err := d.db.Exec(ctx, "call orgs.delete_group($1)", name)
	return err
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

// GetGroupsOfResources returns all the resources group names
func (d *DbStorage) GetGroupsOfResources(ctx context.Context) ([]string, error) {
	var result []string
	if rows, err := d.db.Query(ctx, "select distinct group_name from auth.resources order by group_name asc"); err != nil {
		return result, err
	} else if rows == nil {
		return result, nil
	} else {
		defer rows.Close()

		for rows.Next() {
			if rows.Err() != nil {
				return result, err
			}

			var group string
			if err := rows.Scan(&group); err != nil {
				return result, err
			} else {
				result = append(result, group)
			}
		}
	}

	return result, nil
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

// GetUserGrantAccessPerGroup returns, for each resources group, all roles for that group that the user was granted
func (d DbStorage) GetUserGrantAccessPerGroup(ctx context.Context, username string) (map[string][]dto.GrantRole, error) {
	result := make(map[string][]dto.GrantRole)
	if rows, err := d.db.Query(ctx, "select group_name, roles from auth.get_groups_auths_for_user($1)", username); err != nil {
		return result, err
	} else if rows == nil {
		return result, nil
	} else {
		defer rows.Close()

		for rows.Next() {
			if rows.Err() != nil {
				return result, err
			}

			var group string
			roles := []string{}
			if err := rows.Scan(&group, &roles); err != nil {
				return result, err
			} else if parsedRoles, err := dto.ParseGrantRoles(roles); err != nil {
				return result, err
			} else {
				result[group] = parsedRoles
			}
		}
	}

	return result, nil
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

func (d DbStorage) GrantAccessBatch(ctx context.Context, username string, access map[string][]dto.GrantRole) error {
	if transaction, err := d.db.Begin(ctx); err != nil {
		return err
	} else {
		for group, roles := range access {
			if len(roles) != 0 {
				// grant access
				mapping := make([]string, len(roles))
				for index, value := range roles {
					mapping[index] = string(value)
				}

				if _, err := d.db.Exec(ctx, "call  auth.grant_group_access_to_user($1,$2,$3)", username, mapping, group); err != nil {
					transaction.Rollback(ctx)
					return err
				}

			} else if _, err := transaction.Exec(ctx, "call auth.remove_access_from_group_to_user($1,$2)", username, group); err != nil {
				transaction.Rollback(ctx)
				return err
			}
		}

		if err := transaction.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}

// RemoveAccessToGroupOfResources removes access rights for that user to a given group of resources
func (d DbStorage) RemoveAccessToGroupOfResources(ctx context.Context, username string, group string) error {
	_, err := d.db.Exec(ctx, "call auth.remove_access_from_group_to_user($1,$2)", username, group)
	return err
}

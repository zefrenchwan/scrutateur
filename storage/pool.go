package storage

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func CreateConnectionsPool(config *pgxpool.Config) (*pgxpool.Pool, error) {
	return pgxpool.NewWithConfig(context.Background(), config)
}

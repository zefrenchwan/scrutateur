package main

import (
	"os"
	"time"

	"github.com/zefrenchwan/scrutateur.git/engines"
	"github.com/zefrenchwan/scrutateur.git/services"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

func main() {
	logPath := "logs/server.log"

	///////////////////
	// Init log system
	logger := engines.NewLogger(logPath)

	////////////////////////
	// Launch storage system
	options := storage.DaoOptions{PostgresqlURL: os.Getenv("POSTGRESQL_URL")}
	// Create DAO
	var dao storage.Dao
	if db, err := storage.NewDao(options, logger); err != nil {
		panic(err)
	} else {
		dao = db
	}

	defer dao.Close()

	///////////////////////////////
	// define web serving and links

	secret := os.Getenv("ENGINE_SECRET")
	if secret == "" {
		secret = engines.NewLongSecret()
	}

	engine := services.Init(dao, secret, 24*time.Hour)
	logger.Println("Starting engine")
	// start engine
	engine.Launch(":3000")
}

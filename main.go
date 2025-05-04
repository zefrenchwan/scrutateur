package main

import (
	"os"

	"github.com/zefrenchwan/scrutateur.git/services"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

func main() {
	options := storage.DaoOptions{PostgresqlURL: os.Getenv("POSTGRESQL_URL"), RedisURL: os.Getenv("REDIS_URL")}
	// Create DAO
	var dao storage.Dao
	if db, err := storage.NewDao(options); err != nil {
		panic(err)
	} else {
		dao = db
	}

	defer dao.Close()

	// define web serving and links
	engine := services.NewServer(dao)
	engine.Init()
	// start engine
	engine.Run(3000)
}

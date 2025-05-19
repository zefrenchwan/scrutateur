package main

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/services"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

func main() {
	logPath := "logs/server.log"

	// Init log system
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	var logFile io.Writer
	os.Remove(logPath)
	if file, err := os.Create(logPath); err != nil {
		panic(err)
	} else {
		logFile = io.MultiWriter(file)
	}

	gin.DefaultWriter = logFile
	logger := log.New(logFile, "", log.Ldate|log.Ltime|log.Llongfile)

	options := storage.DaoOptions{PostgresqlURL: os.Getenv("POSTGRESQL_URL"), RedisURL: os.Getenv("REDIS_URL")}
	// Create DAO
	var dao storage.Dao
	if db, err := storage.NewDao(options, logger); err != nil {
		panic(err)
	} else {
		dao = db
	}

	defer dao.Close()

	// define web serving and links
	engine := services.NewServer(dao, logger)
	engine.Init()
	logger.Println("Starting engine")
	// start engine
	engine.Run(3000)
}

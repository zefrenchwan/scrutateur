package main

import (
	"github.com/zefrenchwan/scrutateur.git/services"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

func main() {
	// Create DAO
	var dao storage.Dao
	if db, err := storage.NewDao("postgres", "popo"); err != nil {
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

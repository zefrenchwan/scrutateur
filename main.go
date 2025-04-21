package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

func main() {
	engine := gin.New()
	engine.GET("/test", func(context *gin.Context) {
		if db, err := storage.CreateDatabase("postgres", "popo"); err != nil {
			fmt.Println(err)
		} else {
			var counter int64
			db.Table("auth.users").Count(&counter)
			fmt.Println(counter)
		}
		context.JSON(http.StatusAccepted, "Alors, euh... Bonchouriiin")
	})
	engine.Run(":3000")
}

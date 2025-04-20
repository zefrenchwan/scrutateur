package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.New()
	engine.GET("/test", func(context *gin.Context) {
		context.JSON(http.StatusAccepted, "Alors, euh... Bonchouriiin")
	})
	engine.Run(":3000")
}

package services

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

type Server struct {
	dao    storage.Dao
	engine *gin.Engine
}

func NewServer(dao storage.Dao) Server {
	engine := gin.New()
	return Server{dao: dao, engine: engine}
}

func (s *Server) Status(context *gin.Context) {
	context.JSON(200, nil)
}

func (s *Server) Init() {
	s.engine.GET("/status", func(c *gin.Context) {
		s.Status(c)
	})
}

func (s *Server) Run(port int) {
	s.engine.Run(":" + strconv.Itoa(port))
}

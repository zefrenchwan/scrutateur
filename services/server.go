package services

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

// Server encapsulates app server toolbox (dao) and decorates a gin server
type Server struct {
	secret        string
	tokenDuration time.Duration
	dao           storage.Dao
	engine        *gin.Engine
}

// NewServer makes a new app server with a given dao
func NewServer(dao storage.Dao) Server {
	return Server{dao: dao, engine: gin.New(), secret: NewSecret(), tokenDuration: time.Hour * 24}
}

// Status is a ping handler
func (s *Server) Status(context *gin.Context) {
	context.String(200, "OK")
}

// Init is the place to add all links endpoint -> handlers
func (s *Server) Init() {
	s.engine.GET("/status", func(c *gin.Context) {
		s.Status(c)
	})

	s.engine.POST("/login", func(c *gin.Context) {
		s.Login(c)
	})
}

// Run starts the server
func (s *Server) Run(port int) {
	s.engine.Run(":" + strconv.Itoa(port))
}

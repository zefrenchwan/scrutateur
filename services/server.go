package services

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

// Server encapsulates app server toolbox (dao) and decorates a gin server
type Server struct {
	// secret to deal with auth
	secret string
	// tokenDuration is defaulted to 24h
	tokenDuration time.Duration
	// dao allows any database operation
	dao storage.Dao
	// engine is the technical solution to implement server logic (gin in this case)
	engine *gin.Engine
	// cookieName is the name of the cookie to look for
	cookieName string
}

// NewServer makes a new app server with a given dao
func NewServer(dao storage.Dao) Server {
	return Server{
		dao:           dao,
		engine:        gin.New(),
		secret:        NewSecret(),
		tokenDuration: time.Hour * 24,
		cookieName:    "app.cookie.scrutateur",
	}
}

// Status is a ping handler
func (s *Server) Status(context *gin.Context) {
	context.String(200, "OK")
}

// Init is the place to add all links endpoint -> handlers
func (s *Server) Init() {
	// define useful unprotected functions here
	// status is a ping method
	s.engine.GET("/status", func(c *gin.Context) {
		s.Status(c)
	})

	// login is the connection handler
	s.engine.POST("/login", func(c *gin.Context) {
		s.Login(c)
	})
}

// Run starts the server
func (s *Server) Run(port int) {
	s.engine.Run(":" + strconv.Itoa(port))
}

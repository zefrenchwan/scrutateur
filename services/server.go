package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

// Server encapsulates app server toolbox (dao) and decorates a gin server
type Server struct {
	dao    storage.Dao
	engine *gin.Engine
}

// NewServer makes a new app server with a given dao
func NewServer(dao storage.Dao) Server {
	engine := gin.New()
	return Server{dao: dao, engine: engine}
}

// Status is a ping handler
func (s *Server) Status(context *gin.Context) {
	context.String(200, "OK")
}

// Login tests a POST content (username, password) and validates an user
func (s *Server) Login(context *gin.Context) {
	// used once, defining the user connection
	type UserAuth struct {
		Login    string `form:"login" binding:"required"`
		Password string `form:"password" binding:"required"`
	}
	var auth UserAuth
	if err := context.BindJSON(&auth); err != nil {
		context.String(http.StatusBadRequest, "expecting login and password")
		return
	}

	// validate user auth
	if valid, err := s.dao.ValidateUser(auth.Login, auth.Password); err != nil {
		context.String(http.StatusInternalServerError, "Internal error")
	} else if !valid {
		context.String(http.StatusUnauthorized, "Authentication failure")
	}

	// user auth is valid
	context.JSON(http.StatusAccepted, "Hello")
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

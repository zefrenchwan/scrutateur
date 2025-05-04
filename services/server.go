package services

import (
	"context"
	"net/http"
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
}

// NewServer makes a new app server with a given dao
func NewServer(dao storage.Dao) Server {
	return Server{
		dao:           dao,
		engine:        gin.New(),
		secret:        NewSecret(),
		tokenDuration: time.Hour * 24,
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

	// PROTECTED PAGES
	middleware := s.AuthenticationMiddleware()

	// PAGES FOR AT LEAST A ROLE
	allAuthUsersMiddleware := s.RolesBasedMiddleware(HasARoleCondition())
	s.engine.GET("/user/whoami", middleware, allAuthUsersMiddleware, func(c *gin.Context) {
		var session Session
		if s, err := s.SessionLoad(c); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			session = s
		}

		c.String(http.StatusAccepted, session.CurrentUser)
		c.Next()
	})
}

// Run starts the server
func (s *Server) Run(port int) {
	s.engine.Run(":" + strconv.Itoa(port))
}

// SessionLoad loads session for that context
func (s *Server) SessionLoad(c *gin.Context) (Session, error) {
	sessionId := c.Request.Header.Get("session-id")
	var session Session
	if v, err := s.dao.GetSessionForUser(context.Background(), sessionId); err != nil {
		return session, err
	} else if s, err := SessionLoad(v); err != nil {
		return session, err
	} else {
		return s, nil
	}
}

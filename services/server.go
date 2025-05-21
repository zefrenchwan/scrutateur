package services

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/dto"
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
	// logger displays logs into log output
	logger *log.Logger
}

// NewServer makes a new app server with a given dao
func NewServer(dao storage.Dao, logger *log.Logger) Server {
	return Server{
		dao:           dao,
		engine:        gin.New(),
		secret:        NewSecret(),
		tokenDuration: time.Hour * 24,
		logger:        logger,
	}
}

// Log writes log
func (s *Server) Log(values ...any) {
	s.logger.Println(values...)
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
	allAuthUsersMiddleware := s.RolesBasedMiddleware()

	//////////////////////////////////////////////////////////////////////////////
	// GROUP SELF: USERS GET THEIR OWN INFORMATION OR CHANGE THEIR OWN PASSWORD //
	//////////////////////////////////////////////////////////////////////////////
	s.engine.GET("/user/whoami", middleware, allAuthUsersMiddleware, func(c *gin.Context) {
		s.endpointUserInformation(c)
	})

	s.engine.POST("/user/password", middleware, allAuthUsersMiddleware, func(c *gin.Context) {
		s.endpointChangePassword(c)
	})

	/////////////////////////////////////////////
	// GROUP MANAGEMENT: DEAL WITH USER ACCESS //
	/////////////////////////////////////////////
	s.engine.POST("/admin/user/create", middleware, allAuthUsersMiddleware, func(c *gin.Context) {
		s.endpointAdminCreateUser(c)
	})

	s.engine.DELETE("/root/user/delete/:username", middleware, allAuthUsersMiddleware, func(c *gin.Context) {
		s.endpointRootDeleteUser(c)
	})

	s.engine.GET("/admin/user/roles/:username", middleware, allAuthUsersMiddleware, func(c *gin.Context) {
		s.endpointAdminUserRolesPerGroup(c)
	})
}

// Run starts the server
func (s *Server) Run(port int) {
	s.engine.Run(":" + strconv.Itoa(port))
}

// SessionLoad loads session for that context
func (s *Server) SessionLoad(c *gin.Context) (dto.Session, error) {
	sessionId := c.Request.Header.Get("session-id")
	return s.dao.GetSessionForUser(context.Background(), sessionId)
}

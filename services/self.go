package services

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// endpointUserInformation returns user name for that user
func (s *Server) endpointUserInformation(c *gin.Context) {
	if session, err := s.SessionLoad(c); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else {
		c.String(http.StatusOK, session.CurrentUser)
		c.Next()
	}
}

// endpointChangePassword changes current user's password
func (s *Server) endpointChangePassword(c *gin.Context) {
	if session, err := s.SessionLoad(c); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else if password, err := io.ReadAll(c.Request.Body); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else if !ValidateUserpasswordFormat(string(password)) {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid password format"))
		c.Abort()
	} else if err := s.dao.UpsertUser(context.Background(), session.CurrentUser, string(password)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

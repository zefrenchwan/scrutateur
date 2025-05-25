package services

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// endpointUserInformation returns user name for that user
func (s *Server) endpointUserInformation(c *gin.Context) {
	if value, found := c.Get("login"); !found {
		c.AbortWithError(http.StatusInternalServerError, errors.New("no user found"))
	} else {
		c.String(http.StatusOK, value.(string))
	}
}

// endpointChangePassword changes current user's password
func (s *Server) endpointChangePassword(c *gin.Context) {
	if login, found := c.Get("login"); !found {
		c.AbortWithError(http.StatusInternalServerError, errors.New("no user found"))
	} else if password, err := io.ReadAll(c.Request.Body); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else if !ValidateUserpasswordFormat(string(password)) {
		c.AbortWithError(http.StatusForbidden, errors.New("invalid password format"))
	} else if err := s.dao.UpsertUser(context.Background(), login.(string), string(password)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// endpointRootDeleteUser reads user's login parameter and delete that user (cannot be current user)
func (s *Server) endpointRootDeleteUser(c *gin.Context) {
	username := c.Param("username")
	if len(username) == 0 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("missing username for user deletion"))
		c.Abort()
	} else if !ValidateUsernameFormat(username) {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid username format"))
		c.Abort()
	} else if session, err := s.SessionLoad(c); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else if session.CurrentUser == username {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("cannot delete your own account"))
		c.Abort()
	} else if err := s.dao.DeleteUser(context.Background(), username); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

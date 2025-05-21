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
	} else if !ValidateUsernameFormat(username) {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid username format"))
	} else if currentUser, found := c.Get("login"); !found {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot find user"))
	} else if currentUser.(string) == username {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("cannot delete your own account"))
	} else if err := s.dao.DeleteUser(context.Background(), username); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

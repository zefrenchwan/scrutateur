package services

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserCreationContent is the json data definition to create an user
type UserCreationContent struct {
	Username string `json:"name"`
	Password string `json:"password"`
}

// endpointAdminCreateUser creates an user with no role
func (s *Server) endpointAdminCreateUser(c *gin.Context) {
	var content UserCreationContent
	if err := c.ShouldBindBodyWithJSON(&content); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		c.Abort()
	} else if err := s.dao.UpsertUser(context.Background(), content.Username, content.Password); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		c.Abort()
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

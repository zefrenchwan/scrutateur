package services

import (
	"context"
	"fmt"
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
	} else if !ValidateUsernameFormat(content.Username) {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid username format"))
	} else if !ValidateUserpasswordFormat(content.Password) {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid password format"))
	} else if err := s.dao.UpsertUser(context.Background(), content.Username, content.Password); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

// endpointAdminUserRolesPerGroup displays user information for allocated groups and matching roles
func (s *Server) endpointAdminUserRolesPerGroup(c *gin.Context) {
	username := c.Param("username")
	if len(username) == 0 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("missing username for user roles information"))
	} else if !ValidateUsernameFormat(username) {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid username format"))
	} else if values, err := s.dao.GetUserGrantAccessPerGroup(context.Background(), username); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else if len(values) == 0 {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("no matching user for %s", username))
	} else {
		c.JSON(http.StatusOK, values)
		c.Next()
	}
}

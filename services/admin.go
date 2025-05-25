package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zefrenchwan/scrutateur.git/dto"
)

// UserCreationContent is the json data definition to create an user
type UserCreationContent struct {
	Username string `json:"name"`
	Password string `json:"password"`
}

// endpointAdminCreateUser creates an user with no role
func (s *Server) endpointAdminCreateUser(c *gin.Context) {
	defer c.Request.Body.Close()
	var content UserCreationContent
	if err := c.ShouldBindBodyWithJSON(&content); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	} else if !ValidateUsernameFormat(content.Username) {
		c.AbortWithError(http.StatusForbidden, errors.New("invalid username format"))
	} else if !ValidateUserpasswordFormat(content.Password) {
		c.AbortWithError(http.StatusForbidden, errors.New("invalid password format"))
	} else if err := s.dao.UpsertUser(context.Background(), content.Username, content.Password); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

// endpointRootDeleteUser reads user's login parameter and delete that user (cannot be current user)
func (s *Server) endpointRootDeleteUser(c *gin.Context) {
	username := c.Param("username")
	if len(username) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("missing username for user deletion"))
	} else if !ValidateUsernameFormat(username) {
		c.AbortWithError(http.StatusForbidden, errors.New("invalid username format"))
	} else if currentUser, found := c.Get("login"); !found {
		c.AbortWithError(http.StatusInternalServerError, errors.New("cannot find user"))
	} else if currentUser.(string) == username {
		c.AbortWithError(http.StatusBadRequest, errors.New("cannot delete your own account"))
	} else if err := s.dao.DeleteUser(context.Background(), username); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Status(http.StatusOK)
		c.Next()
	}
}

// endpointAdminListUserRoles displays user information for allocated groups and matching roles
func (s *Server) endpointAdminListUserRoles(c *gin.Context) {
	username := c.Param("username")
	if len(username) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("missing username for user roles information"))
	} else if !ValidateUsernameFormat(username) {
		c.AbortWithError(http.StatusForbidden, errors.New("invalid username format"))
	} else if values, err := s.dao.GetUserGrantAccessPerGroup(context.Background(), username); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else if len(values) == 0 {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("no matching user for %s", username))
	} else {
		c.JSON(http.StatusOK, values)
		c.Next()
	}
}

// endpointAdminEditUserRoles changes roles of a given user for a given group
func (s *Server) endpointAdminEditUserRoles(c *gin.Context) {
	defer c.Request.Body.Close()
	username := c.Param("username")
	var values map[string][]string
	if len(username) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("missing username for user roles information"))
	} else if !ValidateUsernameFormat(username) {
		c.AbortWithError(http.StatusForbidden, errors.New("invalid username format"))
	} else if actor, found := c.Get("login"); !found {
		c.AbortWithError(http.StatusInternalServerError, errors.New("cannot access login from current content"))
	} else if actorAccess, err := s.dao.GetUserGrantAccessPerGroup(context.Background(), actor.(string)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else if err := c.ShouldBindJSON(&values); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	} else if len(values) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("empty request"))
	} else {
		parsedRequest := make(map[string][]dto.GrantRole)
		// request may be parsed
		for group, rawRoles := range values {
			if len(group) == 0 {
				c.AbortWithError(http.StatusBadRequest, errors.New("empty value"))
				return
			} else if roles, err := dto.ParseGrantRoles(rawRoles); err != nil {
				c.AbortWithError(http.StatusBadRequest, errors.New("invalid roles"))
				return
			} else {
				parsedRequest[group] = roles
			}
		}

		if err := MayGrant(actorAccess, parsedRequest); err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
		} else if err := s.dao.GrantAccess(context.Background(), username, parsedRequest); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			c.Status(http.StatusOK)
			c.Next()
		}
	}
}

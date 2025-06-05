package engines

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/dto"
)

// EndpointAdminCreateUser creates an user with no role
func EndpointAdminCreateUser(c *HandlerContext) error {
	var headers http.Header
	var content UserInformation
	if err := c.BindJsonBody(&content); err != nil {
		c.BuildError(http.StatusBadRequest, err, headers)
	} else if !ValidateUsernameFormat(content.Username) {
		c.Build(http.StatusForbidden, "invalid username format", headers)
	} else if !ValidateUserpasswordFormat(content.Password) {
		c.Build(http.StatusForbidden, "invalid password format", headers)
	} else if err := c.Dao.UpsertUser(context.Background(), content.Username, content.Password); err != nil {
		c.BuildError(http.StatusInternalServerError, err, headers)
	} else {
		headers = c.RequestHeaderByNames("Authorization")
		c.Build(http.StatusOK, "", headers)
	}

	// no error we could not manage
	return nil
}

// EndpointRootDeleteUser reads user's login parameter and delete that user (cannot be current user)
func EndpointRootDeleteUser(c *HandlerContext) error {
	username := c.GetQueryParameters()["username"]
	var headers http.Header
	if len(username) == 0 {
		c.Build(http.StatusBadRequest, "missing username for user deletion", headers)
	} else if !ValidateUsernameFormat(username) {
		c.Build(http.StatusForbidden, "invalid username format", nil)
	} else if currentUser := c.GetContextValueAsString("login"); currentUser == "" {
		c.Build(http.StatusInternalServerError, "cannot find user", nil)
	} else if currentUser == username {
		c.Build(http.StatusBadRequest, "cannot delete your own account", nil)
	} else if err := c.Dao.DeleteUser(context.Background(), username); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
	} else {
		c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
	}

	// dealt with internal errors already
	return nil
}

// EndpointAdminListUserRoles displays user information for allocated groups and matching roles
func EndpointAdminListUserRoles(c *HandlerContext) error {
	username := c.GetQueryParameters()["username"]
	if len(username) == 0 {
		c.Build(http.StatusBadRequest, "missing username for user roles information", nil)
	} else if !ValidateUsernameFormat(username) {
		c.Build(http.StatusForbidden, "invalid username format", nil)
	} else if values, err := c.Dao.GetUserGrantAccessPerGroup(context.Background(), username); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
	} else if len(values) == 0 {
		c.Build(http.StatusNotFound, fmt.Sprintf("no matching user for %s", username), nil)
	} else if err := c.BuildJson(http.StatusOK, values, c.RequestHeaderByNames("Authorization")); err != nil {
		c.ClearResponse()
		c.BuildError(http.StatusInternalServerError, err, nil)
	}

	// no unmanageable error
	return nil
}

// EndpointAdminEditUserRoles changes roles of a given user for a given group
func EndpointAdminEditUserRoles(c *HandlerContext) error {
	username := c.GetQueryParameters()["username"]
	var values map[string][]string
	if len(username) == 0 {
		c.Build(http.StatusBadRequest, "missing username for user roles information", nil)
	} else if !ValidateUsernameFormat(username) {
		c.Build(http.StatusForbidden, "invalid username format", nil)
	} else if actor := c.GetContextValueAsString("login"); actor == "" {
		c.Build(http.StatusInternalServerError, "cannot access login from current content", nil)
	} else if actorAccess, err := c.Dao.GetUserGrantAccessPerGroup(context.Background(), actor); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
	} else if err := c.BindJsonBody(&values); err != nil {
		c.BuildError(http.StatusBadRequest, err, nil)
	} else if len(values) == 0 {
		c.Build(http.StatusBadRequest, "empty request", nil)
	} else {
		parsedRequest := make(map[string][]dto.GrantRole)
		// request may be parsed
		for group, rawRoles := range values {
			if len(group) == 0 {
				c.Build(http.StatusBadRequest, "empty value", nil)
				return nil
			} else if roles, err := dto.ParseGrantRoles(rawRoles); err != nil {
				c.Build(http.StatusBadRequest, "invalid roles", nil)
				return nil
			} else {
				parsedRequest[group] = roles
			}
		}

		if err := MayGrant(actorAccess, parsedRequest); err != nil {
			c.BuildError(http.StatusUnauthorized, err, nil)
		} else if err := c.Dao.GrantAccess(context.Background(), username, parsedRequest); err != nil {
			c.BuildError(http.StatusInternalServerError, err, nil)
		} else {
			c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
		}
	}

	return nil
}

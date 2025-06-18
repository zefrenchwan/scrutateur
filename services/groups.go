package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"

	"github.com/zefrenchwan/scrutateur.git/dto"
	"github.com/zefrenchwan/scrutateur.git/engines"
)

// endpointCreateGroup creates a group by name, or raise an error
func endpointCreateGroup(c *engines.HandlerContext) error {
	name := c.GetQueryParameters()["groupName"]
	if name == "" {
		c.Build(http.StatusInternalServerError, "missing group parameter", nil)
		return nil
	} else if !ValidateGroupNameFormat(name) {
		c.Build(http.StatusBadRequest, "group parameter does not match valid group name rules", nil)
		return nil
	}

	creator := c.GetLogin()
	groupRoles := append([]dto.GrantRole{dto.RoleReader, dto.RoleEditor}, c.GetRoles()...)
	groupRoles = slices.Compact(groupRoles)

	if err := c.Dao.CreateUsersGroup(context.Background(), creator, name, groupRoles); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
		return nil
	} else {
		c.Dao.LogEvent(context.Background(), creator, "groups", fmt.Sprintf("user %s creates group %s", creator, name), nil)
		c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
		return nil
	}
}

// endpointListGroupsForUser displays groups an user is in, with access rights
func endpointListGroupsForUser(c *engines.HandlerContext) error {
	login := c.GetLogin()
	if login == "" {
		c.Build(http.StatusUnauthorized, "no active user", nil)
		return nil
	}

	if values, err := c.Dao.ListUserGroupsForSpecificUser(context.Background(), login); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
		return nil
	} else if len(values) == 0 {
		c.Build(http.StatusNoContent, "", nil)
		return nil
	} else if err := c.BuildJson(http.StatusOK, values, c.RequestHeaderByNames("Authorization")); err != nil {
		c.ClearResponse()
		c.BuildError(http.StatusInternalServerError, err, nil)
		return nil
	} else {
		return nil
	}
}

// endpointUpsertUserInGroup allows to change user (or add user) within a group
func endpointUpsertUserInGroup(c *engines.HandlerContext) error {
	parameters := c.GetQueryParameters()
	groupName := parameters["groupName"]
	userName := parameters["userName"]
	login := c.GetLogin()

	if !engines.ValidateUsernameFormat(userName) {
		c.Build(http.StatusBadRequest, "invalid user format", nil)
		return nil
	} else if !ValidateGroupNameFormat(groupName) {
		c.Build(http.StatusBadRequest, "invalid group format", nil)
		return nil
	}

	// Read body, expect list of roles
	var body []byte
	if groupName == "" {
		c.Build(http.StatusInternalServerError, "missing group parameter", nil)
		return nil
	} else if !ValidateGroupNameFormat(groupName) {
		c.Build(http.StatusBadRequest, "group parameter does not match valid group name rules", nil)
		return nil
	} else if raw, err := c.RequestBodyAsString(); err != nil {
		c.BuildError(http.StatusBadRequest, err, nil)
		return nil
	} else if regexp.MustCompile(`\A\s*\z`).MatchString(raw) {
		c.Build(http.StatusBadRequest, "empty auth, need at least one", nil)
		return nil
	} else {
		body = []byte(raw)
	}

	var values []any
	var request []dto.GrantRole
	if err := json.Unmarshal(body, &values); err != nil {
		unmarshallError := errors.Join(errors.New("error when unmarshalling: "+string(body)), err)
		c.BuildError(http.StatusBadRequest, unmarshallError, nil)
		return nil
	} else {
		for _, val := range values {
			if val == nil {
				continue
			} else if r, ok := val.(string); !ok {
				c.Build(http.StatusBadRequest, "cannot read roles", nil)
				return nil
			} else if role, err := dto.ParseGrantRole(r); err != nil {
				c.Build(http.StatusBadRequest, "cannot read roles", nil)
				return nil
			} else {
				request = append(request, role)
			}
		}
	}

	// Now, ensure that roles match and insert
	globalRoles := c.GetRoles()
	localRoles, errLoad := c.Dao.GetGroupAuthForUser(c.GetCurrentContext(), login, groupName)
	if errLoad != nil {
		c.BuildError(http.StatusInternalServerError, errLoad, nil)
		return nil
	} else if len(request) == 0 {
		c.Build(http.StatusBadRequest, "empty auth, need at least one", nil)
		return nil
	} else if !HasMinimumAccessAuth(globalRoles, localRoles, request) {
		c.Build(http.StatusUnauthorized, "insufficient privilege for user "+userName, nil)
		return nil
	} else if err := c.Dao.SetGroupAuthForUser(c.GetCurrentContext(), login, userName, groupName, request); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
		return nil
	}

	var paramRoles []string
	for _, role := range request {
		paramRoles = append(paramRoles, string(role))
	}

	c.Dao.LogEvent(c.GetCurrentContext(), login, "groups", fmt.Sprintf("user %s upserts user %s within group %s", login, userName, groupName), paramRoles)
	c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
	return nil
}

// endpointRevokeUserInGroup revokes a given user within a group
func endpointRevokeUserInGroup(c *engines.HandlerContext) error {
	parameters := c.GetQueryParameters()
	groupName := parameters["groupName"]
	userName := parameters["userName"]
	login := c.GetLogin()

	if !engines.ValidateUsernameFormat(userName) {
		c.Build(http.StatusBadRequest, "invalid user format", nil)
		return nil
	} else if !ValidateGroupNameFormat(groupName) {
		c.Build(http.StatusBadRequest, "invalid group format", nil)
		return nil
	}

	globalRoles := c.GetRoles()
	localRoles, errLoad := c.Dao.GetGroupAuthForUser(c.GetCurrentContext(), login, groupName)
	expectedRoles := []dto.GrantRole{dto.RoleAdmin, dto.RoleEditor, dto.RoleRoot}
	if errLoad != nil {
		c.BuildError(http.StatusInternalServerError, errLoad, nil)
		return nil
	} else if !HasMinimumAccessAuth(globalRoles, localRoles, expectedRoles) {
		c.Build(http.StatusUnauthorized, "insufficient privileges to revoke user", nil)
		return nil
	} else if err := c.Dao.RevokeUserInGroup(c.GetCurrentContext(), userName, groupName); err != nil {
		c.BuildError(http.StatusInternalServerError, errLoad, nil)
		return nil
	}

	c.Dao.LogEvent(c.GetCurrentContext(), login, "groups", fmt.Sprintf("user %s removes user %s from group %s", login, userName, groupName), nil)
	c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
	return nil

}

// endpointDeleteGroup deletes a group by name
func endpointDeleteGroup(c *engines.HandlerContext) error {
	name := c.GetQueryParameters()["groupName"]
	if name == "" {
		c.Build(http.StatusInternalServerError, "missing group parameter", nil)
		return nil
	} else if !ValidateGroupNameFormat(name) {
		c.Build(http.StatusBadRequest, "group parameter does not match valid group name rules", nil)
		return nil
	}

	mandatoryRoles := []dto.GrantRole{dto.RoleAdmin, dto.RoleRoot}
	// global role is necessary to perform that action.
	// But user may have local access rights to consider too
	if localRoles, err := c.Dao.GetGroupAuthForUser(c.GetCurrentContext(), c.GetLogin(), name); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
		return nil
	} else if !HasMinimumAccessAuth(c.GetRoles(), localRoles, mandatoryRoles) {
		c.Build(http.StatusUnauthorized, "insufficient role or group auth", nil)
		return nil
	}

	if err := c.Dao.DeleteUsersGroup(context.Background(), name); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
		return nil
	} else {
		login := c.GetLogin()
		c.Dao.LogEvent(context.Background(), login, "groups", fmt.Sprintf("user %s deletes group %s", login, name), nil)
		c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
		return nil
	}
}

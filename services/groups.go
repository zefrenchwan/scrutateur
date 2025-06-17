package services

import (
	"context"
	"fmt"
	"net/http"
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

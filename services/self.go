package services

import (
	"context"
	"net/http"

	"github.com/zefrenchwan/scrutateur.git/engines"
)

// endpointUserInformation returns user name for that user
func endpointUserInformation(c *engines.HandlerContext) error {
	if value := c.GetContextValueAsString("login"); value == "" {
		c.Build(http.StatusInternalServerError, "no user found", nil)
	} else {
		c.Build(http.StatusOK, value, nil)
	}

	// no unprocessable error
	return nil
}

// endpointChangePassword changes current user's password
func endpointChangePassword(c *engines.HandlerContext) error {
	if login := c.GetContextValueAsString("login"); login == "" {
		c.Build(http.StatusInternalServerError, "no user found", nil)
	} else if password, err := c.RequestBodyAsString(); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
	} else if !ValidateUserpasswordFormat(string(password)) {
		c.Build(http.StatusForbidden, "invalid password format", nil)
	} else if err := c.Dao.UpsertUser(context.Background(), login, password); err != nil {
		c.BuildError(http.StatusInternalServerError, err, nil)
	} else {
		c.Build(http.StatusOK, "", c.RequestHeaderByNames("Authorization"))
	}

	// no unprocessable error
	return nil
}

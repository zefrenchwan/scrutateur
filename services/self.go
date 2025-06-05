package services

import (
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

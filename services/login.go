package services

import (
	"net/http"
	"time"

	"github.com/zefrenchwan/scrutateur.git/engines"
)

// Login tests a POST content (username, password) and validates an user
func BuildLoginHandler(secret string, tokenDuration time.Duration) engines.RequestProcessor {
	return func(c *engines.HandlerContext) error {
		var auth UserInformation
		if err := c.BindJsonBody(&auth); err != nil {
			c.Build(http.StatusBadRequest, "expecting name and password", nil)
			return nil
		} else if valid, err := c.Dao.ValidateUser(c.GetCurrentContext(), auth.Username, auth.Password); err != nil {
			// validate user auth
			c.BuildError(http.StatusInternalServerError, err, nil)
			return nil
		} else if !valid {
			c.Build(http.StatusUnauthorized, "", nil)
			return nil
		} else if token, err := engines.CreateToken(auth.Username, secret, tokenDuration); err != nil {
			c.BuildError(http.StatusInternalServerError, err, nil)
			return nil
		} else {
			// user auth is valid
			c.SetResponseHeader("Authorization", "Bearer "+token)
			c.SetResponseStatus(http.StatusAccepted)
			c.Done()
			return nil
		}
	}
}

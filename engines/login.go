package engines

import (
	"context"
	"net/http"
	"time"
)

// Login tests a POST content (username, password) and validates an user
func BuildLoginHandler(secret string, tokenDuration time.Duration) RequestProcessor {
	return func(c *HandlerContext) error {
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
		} else if token, err := CreateToken(auth.Username, secret, tokenDuration); err != nil {
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

// EndpointChangePassword changes current user's password
func EndpointChangePassword(c *HandlerContext) error {
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

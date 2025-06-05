package engines

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ValidateQueryProcessor returns a processor that validates the query method type
func ValidateQueryProcessor(requestMethod string) RequestProcessor {
	return func(context *HandlerContext) error {
		if !strings.EqualFold(requestMethod, context.GetRequestMethod()) {
			context.Build(http.StatusMethodNotAllowed, "", nil)
		}

		return nil
	}
}

// AuthenticationMiddleware builds a middleware to deal with auth
func AuthenticationMiddleware(secret string, tokenDuration time.Duration) RequestProcessor {
	// this function tests the token, test session and then sets main headers
	return func(c *HandlerContext) error {
		// get the bearer and token as a whole reading the header
		tokenString := c.GetRequestHeaderFirstValue("Authorization")
		if tokenString == "" {
			c.Build(http.StatusUnauthorized, "missing authorization header", nil)
			return nil
		}

		// get rid of "Bearer " to get only the token
		tokenString = tokenString[len("Bearer "):]

		// Either token is valid and we know the user, or we stop right here.
		// If token is valid, renew the token so that user has more time
		if token, err := VerifyToken(secret, tokenString); err != nil {
			c.BuildError(http.StatusUnauthorized, err, nil)
			return nil
		} else if newToken, err := CreateToken(token.Username, secret, tokenDuration); err != nil {
			c.Build(http.StatusInternalServerError, fmt.Sprintf("cannot renew token: %s", err.Error()), nil)
			return nil
		} else {
			c.SetResponseHeader("Authorization", "Bearer "+newToken)
			c.SetContextValue("login", token.Username)
			return nil
		}
	}
}

// RolesBasedMiddleware tests if user may access this page or not, based on roles based conditions in database
func RolesBasedMiddleware() RequestProcessor {
	// this function tests the token, test session and then sets main headers
	return func(c *HandlerContext) error {
		if login := c.GetContextValueAsString("login"); login == "" {
			c.Build(http.StatusInternalServerError, "no user found", nil)
		} else if conditions, err := c.Dao.GetUserGrantedAccess(context.Background(), login); err != nil {
			c.BuildError(http.StatusInternalServerError, err, nil)
		} else {
			engine := AuthRulesEngine{Conditions: conditions}
			if accept, roles, err := engine.CanAccessResource(c.GetRequestPath()); err != nil {
				c.BuildError(http.StatusInternalServerError, err, nil)
			} else if !accept {
				c.Build(http.StatusUnauthorized, "cannot access resource due to missing permissions", nil)
			} else {
				c.SetContextValue("roles", roles)
			}
		}

		// no unprocessable exception
		return nil
	}
}

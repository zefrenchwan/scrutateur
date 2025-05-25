package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zefrenchwan/scrutateur.git/dto"
)

// NewSecret builds a new secret
func NewSecret() string {
	base := uuid.NewString() + uuid.NewString()
	return strings.ReplaceAll(base, "-", "")
}

// Thanks to
// https://medium.com/@cheickzida/golang-implementing-jwt-token-authentication-bba9bfd84d60
// for the JWT token management

// Login tests a POST content (username, password) and validates an user
func (s *Server) Login(c *gin.Context) {
	// used once, defining the user connection
	type UserAuth struct {
		Login    string `form:"login" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	var auth UserAuth
	if err := c.BindJSON(&auth); err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("expecting login and password"))
	} else if valid, err := s.dao.ValidateUser(context.Background(), auth.Login, auth.Password); err != nil {
		// validate user auth
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if token, err := CreateToken(auth.Login, s.secret, s.tokenDuration); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else {
		// user auth is valid
		c.Writer.Header().Add("Authorization", "Bearer "+token)
		c.JSON(http.StatusAccepted, "Hello "+auth.Login)
		c.Next()
	}
}

// CreateToken creates a string token for a given user, based on a secret.
// NOTE THAT secret is not the user's password
// Token is valid for a given duration
func CreateToken(username, secret string, delay time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(delay).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken uses the secret to check for a token and returns either login,nil or "", and error for invalid token
func VerifyToken(secret string, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		user := claims["username"]
		return user.(string), nil
	} else {
		return "", errors.New("unsupported claim type for JWT token")
	}
}

// AuthenticationMiddleware builds a middleware to deal with auth
func (s *Server) AuthenticationMiddleware() gin.HandlerFunc {
	// this function tests the token, test session and then sets main headers
	return func(c *gin.Context) {
		// get the bearer and token as a whole reading the header
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.AbortWithError(http.StatusUnauthorized, errors.New("missing authorization header"))
			return
		}

		// get rid of "Bearer " to get only the token
		tokenString = tokenString[len("Bearer "):]

		// Either token is valid and we know the user, or we stop right here
		if login, err := VerifyToken(s.secret, tokenString); err != nil {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid token: %s", err.Error()))
			return
		} else {
			c.Set("login", login)
		}

		// All security tests passed
		// Set token for security
		// Source: https://gin-gonic.com/en/docs/examples/security-headers/
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")

		// add auth and session
		c.Header("Authorization", "Bearer "+tokenString)

		c.Next()
	}
}

// AuthRulesEngine applies grant conditions and tests whether an user may access a resource
type AuthRulesEngine struct {
	// Conditions to apply
	Conditions []dto.GrantAccessForResource
}

// CanAccessResource returns true and roles for user if user may access, false and nil otherwise. Error if any as the last value
func (re *AuthRulesEngine) CanAccessResource(url string) (bool, []dto.GrantRole, error) {
	regexpValidator := regexp.MustCompile(REGEXP_URL_PART)
	for _, condition := range re.Conditions {
		templateUrl := condition.Template
		expectedRoles := condition.UserRoles
		switch condition.Operator {
		case dto.OperatorEquals:
			if templateUrl == url {
				return true, expectedRoles, nil
			}
		case dto.OperatorStartsWith:
			if strings.HasPrefix(url, templateUrl) {
				return true, expectedRoles, nil
			}
		case dto.OperatorMatches:
			localTest := true
			urlParts := strings.Split(url, "/")
			templateParts := strings.Split(templateUrl, "/")
			size := len(urlParts)

			// test first that there is exactly the same amount of / parts
			if size != len(templateParts) {
				localTest = false
			} else {
				for index, value := range urlParts {
					if !localTest {
						break
					}
					templatePart := templateParts[index]
					if templatePart == "*" {
						localTest = regexpValidator.MatchString(value)
					} else {
						localTest = (value == templatePart)
					}
				}
			}

			if localTest {
				return true, expectedRoles, nil
			}
		}
	}

	return false, nil, nil
}

// RolesBasedMiddleware tests if user may access this page or not, based on roles based conditions in database
func (s *Server) RolesBasedMiddleware() gin.HandlerFunc {
	// this function tests the token, test session and then sets main headers
	return func(c *gin.Context) {
		if login, found := c.Get("login"); !found {
			c.AbortWithError(http.StatusInternalServerError, errors.New("no user found"))
		} else if conditions, err := s.dao.GetUserGrantedAccess(context.Background(), login.(string)); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			engine := AuthRulesEngine{Conditions: conditions}
			if accept, roles, err := engine.CanAccessResource(c.Request.URL.Path); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			} else if !accept {
				c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("cannot access %s due to missing permissions", c.Request.RequestURI))
			} else {
				c.Set("roles", roles)
				c.Next()
			}
		}
	}
}

// MayGrant returns an error if adminAccess ore not sufficient to grant requestedAccess.
// Parameters are group => roles of user
func MayGrant(adminAccess map[string][]dto.GrantRole, requestedAccess map[string][]dto.GrantRole) error {
	if len(adminAccess) == 0 {
		return errors.New("no admin access")
	} else if len(requestedAccess) == 0 {
		return nil
	} else {
		for group, roles := range requestedAccess {
			if len(group) == 0 {
				return errors.New("empty value")
			} else if accessRights, found := adminAccess[group]; !found {
				return fmt.Errorf("cannot grant on group %s", group)
			} else if !slices.Contains(accessRights, dto.RoleRoot) && !slices.Contains(accessRights, dto.RoleAdmin) {
				return errors.New("cannot grant access due to insufficient privileges: needs admin or root")
			} else if slices.Contains(roles, dto.RoleRoot) && !slices.Contains(accessRights, dto.RoleRoot) {
				return errors.New("cannot grant access due to insufficient privileges: admin cannot grant root")
			}
		}
	}

	return nil
}

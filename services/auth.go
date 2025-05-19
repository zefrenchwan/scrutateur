package services

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
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
		c.String(http.StatusBadRequest, "expecting login and password")
		return
	}

	// validate user auth
	if valid, err := s.dao.ValidateUser(context.Background(), auth.Login, auth.Password); err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, "Internal error")
		return
	} else if !valid {
		c.String(http.StatusUnauthorized, "Authentication failure")
		return
	}

	var newToken string
	if token, err := CreateToken(auth.Login, s.secret, s.tokenDuration); err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, "Cannot generate token for user")
		return
	} else {
		newToken = token
	}

	// set session id value
	newSessionId := NewSecret()
	session := dto.NewSessionForUser(auth.Login)
	if err := s.dao.SetSessionForUser(context.Background(), newSessionId, session); err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, "Cannot store session")
		return
	}

	// session creation went fine, so
	c.Header("session-id", newSessionId)

	// user auth is valid
	c.Writer.Header().Add("Authorization", "Bearer "+newToken)
	c.JSON(http.StatusAccepted, "Hello "+auth.Login)
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
		return "", fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		user := claims["username"]
		return user.(string), nil
	} else {
		return "", fmt.Errorf("unsupported claim type for JWT token")
	}
}

// AuthenticationMiddleware builds a middleware to deal with auth
func (s *Server) AuthenticationMiddleware() gin.HandlerFunc {
	// this function tests the token, test session and then sets main headers
	return func(c *gin.Context) {
		// get the bearer and token as a whole reading the header
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("missing authorization header"))
			return
		}

		// get rid of "Bearer " to get only the token
		tokenString = tokenString[len("Bearer "):]

		var username string
		// Either token is valid and we know the user, or we stop right here
		if login, err := VerifyToken(s.secret, tokenString); err != nil {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid token: %s", err.Error()))
			return
		} else {
			username = login
		}

		// TOKEN IS VALID AND USER IS KNOWN
		// Now we want to check user session
		sessionId := c.Request.Header.Get("session-id")
		// check that session id fits that user
		if session, err := s.dao.GetSessionForUser(context.Background(), sessionId); err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("session loading failure: %s", err.Error()))
			c.Abort()
			return
		} else if session.CurrentUser != username {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("session mismatch"))
			c.Abort()
			return
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
		c.Header("session-id", sessionId)

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
		case dto.OperatorContains:
			if strings.Contains(url, templateUrl) {
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
		if session, err := s.SessionLoad(c); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			c.Abort()
		} else if conditions, err := s.dao.GetUserGrantedAccess(context.Background(), session.CurrentUser); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			c.Abort()
		} else {
			engine := AuthRulesEngine{Conditions: conditions}
			if accept, _, err := engine.CanAccessResource(c.Request.URL.Path); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				c.Abort()
			} else if !accept {
				c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("cannot access %s due to missing permissions", c.Request.RequestURI))
				c.Abort()
			} else {
				c.Next()
			}
		}
	}
}

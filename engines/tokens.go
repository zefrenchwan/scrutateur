package engines

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Content of the token when using JWT
type TokenContent struct {
	Username       string
	ExpirationTime time.Time
}

// Thanks to
// https://medium.com/@cheickzida/golang-implementing-jwt-token-authentication-bba9bfd84d60
// for the JWT token management

// CreateToken creates a string token for a given user, based on a secret.
// NOTE THAT secret is not the user's password
// Token is valid for a given duration
func CreateToken(username, secret string, delay time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().UTC().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken uses the secret to check for a token and returns either the token content,nil or empty, and error for invalid token
func VerifyToken(secret string, tokenString string) (TokenContent, error) {
	var content TokenContent
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return content, err
	} else if !token.Valid {
		return content, errors.New("invalid token")
	} else if claims, ok := token.Claims.(jwt.MapClaims); !ok {
		return content, errors.New("unsupported claim type for JWT token")
	} else if timeValue, err := claims.GetExpirationTime(); err != nil {
		return content, errors.New("cannot read token expiration time")
	} else {
		content.ExpirationTime = timeValue.Time
		content.Username = claims["username"].(string)
		return content, nil
	}
}

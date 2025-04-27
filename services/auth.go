package services

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewSecret() string {
	base := uuid.NewString() + uuid.NewString()
	return strings.ReplaceAll(base, "-", "")
}

// Thanks to
// https://medium.com/@cheickzida/golang-implementing-jwt-token-authentication-bba9bfd84d60
// for the JWT token management

// Login tests a POST content (username, password) and validates an user
func (s *Server) Login(context *gin.Context) {
	// used once, defining the user connection
	type UserAuth struct {
		Login    string `form:"login" binding:"required"`
		Password string `form:"password" binding:"required"`
	}
	var auth UserAuth
	if err := context.BindJSON(&auth); err != nil {
		context.String(http.StatusBadRequest, "expecting login and password")
		return
	}

	// don't keep the password, keep its hash
	// Clean auth as soon as possible
	hash := sha256.New()
	hash.Write([]byte(auth.Password))
	hashedPassword := string(hash.Sum(nil))
	userLogin := auth.Login
	auth = UserAuth{}

	// validate user auth
	if valid, err := s.dao.ValidateUser(userLogin, hashedPassword); err != nil {
		context.String(http.StatusInternalServerError, "Internal error")
	} else if !valid {
		context.String(http.StatusUnauthorized, "Authentication failure")
	}

	if newToken, err := CreateToken(userLogin, s.secret, s.tokenDuration); err == nil {
		context.Writer.Header().Add("Authorization", "Bearer "+newToken)
	} else {
		fmt.Println(err)
		context.String(http.StatusInternalServerError, "Cannot generate token for user")
	}

	// user auth is valid
	context.JSON(http.StatusAccepted, "Hello "+userLogin)
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

// VerifyToken uses the secret to check for a token
func VerifyToken(secret string, tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return secret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		fmt.Println(claims["foo"], claims["nbf"])
	} else {
		fmt.Println(err)
	}

	return nil
}

// ProtectedHandler deals with
// func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	tokenString := r.Header.Get("Authorization")
// 	if tokenString == "" {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		fmt.Fprint(w, "Missing authorization header")
// 		return
// 	}

// 	tokenString = tokenString[len("Bearer "):]
// 	if err := VerifyToken(tokenString); err != nil {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		fmt.Fprint(w, "Invalid token")
// 		return
// 	}

// 	fmt.Fprint(w, "Welcome to the the protected area")

// }

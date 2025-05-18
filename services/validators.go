package services

import "regexp"

// ValidateUsernameFormat tests if username format is valid or not
func ValidateUsernameFormat(username string) bool {
	return regexp.MustCompile(`[a-zA-Z]+\w+`).Match([]byte(username))
}

// ValidateUserpasswordFormat tests if user password format is valid or not
func ValidateUserpasswordFormat(password string) bool {
	return regexp.MustCompile(`\S{4,60}`).Match([]byte(password))
}

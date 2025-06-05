package engines

import (
	"regexp"
)

// REGEXP_URL_PART defines what is acceptable for an url part: /part1/part2/part3
const REGEXP_URL_PART = `^[a-zA-Z0-9_\-]+$`

// ValidateUsernameFormat tests if username format is valid or not
func ValidateUsernameFormat(username string) bool {
	if res, err := regexp.MatchString(`^[a-zA-Z]{4,}[0-9]*$`, username); err != nil {
		panic(err)
	} else {
		return res
	}
}

// ValidateUserpasswordFormat tests if user password format is valid or not
func ValidateUserpasswordFormat(password string) bool {
	if res, err := regexp.MatchString(`^\S{4,60}$`, password); err != nil {
		panic(err)
	} else {
		return res
	}
}

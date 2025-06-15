package services

import "regexp"

// ValidateGroupNameFormat tests if group format is valid or not
func ValidateGroupNameFormat(groupName string) bool {
	if res, err := regexp.MatchString(`^[a-zA-Z]+[0-9]*$`, groupName); err != nil {
		panic(err)
	} else {
		return res
	}
}

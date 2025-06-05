package engines

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zefrenchwan/scrutateur.git/dto"
)

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

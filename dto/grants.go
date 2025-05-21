package dto

import "fmt"

/////////////////////////////////////////////////////////////////
// GRANT ROLE IS WHAT SOMEONE MAY DO AND LIMITS TO ITS ACTIONS //
/////////////////////////////////////////////////////////////////

// GrantRole defines roles implemented
type GrantRole string

// Possible values are listed here
const (
	RoleRoot   GrantRole = "root"
	RoleAdmin  GrantRole = "admin"
	RoleEditor GrantRole = "editor"
	RoleReader GrantRole = "reader"
)

// ParseGrantRole gets a string and returns matching role if any, or error
func ParseGrantRole(value string) (GrantRole, error) {
	switch value {
	case string(RoleRoot):
		return RoleRoot, nil
	case string(RoleAdmin):
		return RoleAdmin, nil
	case string(RoleEditor):
		return RoleEditor, nil
	case string(RoleReader):
		return RoleReader, nil
	}

	var empty GrantRole
	return empty, fmt.Errorf("%s is not a grant role", value)
}

// ParseGrantRoles tries to convert each element as a role and returns the mapped array, or error
func ParseGrantRoles(values []string) ([]GrantRole, error) {
	if values == nil {
		return nil, nil
	}

	result := make([]GrantRole, len(values))
	for index, value := range values {
		if role, err := ParseGrantRole(value); err != nil {
			return result, err
		} else {
			result[index] = role
		}
	}

	return result, nil
}

/////////////////////////////////////////////////////////////////
// GRANT OPERATOR IS WHAT OPERATOR TO TEST A TEMPLATE MATCHING //
/////////////////////////////////////////////////////////////////

// GrantOperator is an enum to define operations on a template to validate
type GrantOperator string

// Possible values are listed here
const (
	OperatorEquals     GrantOperator = "EQUALS"
	OperatorStartsWith GrantOperator = "STARTS_WITH"
	OperatorMatches    GrantOperator = "MATCHES"
)

// ParseGrantOperator
func ParseGrantOperator(value string) (GrantOperator, error) {
	switch value {
	case string(OperatorEquals):
		return OperatorEquals, nil
	case string(OperatorStartsWith):
		return OperatorStartsWith, nil
	case string(OperatorMatches):
		return OperatorMatches, nil
	}

	var empty GrantOperator
	return empty, fmt.Errorf("%s is not a grant operator", value)
}

////////////////////////////////////////////
// GRANT ACCESS USER BASED ON A CONDITION //
////////////////////////////////////////////

// GrantAccessForResource is a condition to access a resource for a given user
type GrantAccessForResource struct {
	// Operator defining the resources condition (for instance EQUALS or MATCHES)
	Operator GrantOperator
	// Template is an URL or a regexp based template to accept a group of URL
	Template string
	// UserRoles are the roles this user may impersonate when accessing that page
	UserRoles []GrantRole
}

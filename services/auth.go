package services

import (
	"slices"

	"github.com/zefrenchwan/scrutateur.git/dto"
)

// HasMinimumAccessAuth tests if user has sufficient auth to meet expected roles expectations.
// Formally, it means getting the union of roles and localRoles, and then find if there is one role that is also in expectedRoles
func HasMinimumAccessAuth(roles []dto.GrantRole, localRoles []dto.GrantRole, expectedRoles []dto.GrantRole) bool {
	result := append(roles, localRoles...)
	result = slices.Compact(result)
	for _, expected := range expectedRoles {
		if slices.Contains(result, expected) {
			return true
		}
	}

	return false
}

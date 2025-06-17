package services

import (
	"slices"

	"github.com/zefrenchwan/scrutateur.git/dto"
)

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

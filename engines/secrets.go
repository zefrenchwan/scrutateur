package engines

import (
	"strings"

	"github.com/google/uuid"
)

// NewSecret builds a new secret
func NewSecret() string {
	base := uuid.NewString() + uuid.NewString()
	return strings.ReplaceAll(base, "-", "")
}

// NewLongSecret builds a long secret expected to last long
func NewLongSecret() string {
	base := uuid.NewString() + uuid.NewString() + uuid.NewString() + uuid.NewString()
	return strings.ReplaceAll(base, "-", "")
}

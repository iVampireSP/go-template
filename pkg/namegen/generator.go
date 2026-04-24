package namegen

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Generate generates a random suffix for namespace
// Returns an 8-character lowercase alphanumeric string
func Generate() string {
	raw := strings.ReplaceAll(uuid.NewString(), "-", "")
	suffix := strings.ToLower(raw)
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return suffix
}

// GenerateWithPrefix generates a namespace with the given prefix
func GenerateWithPrefix(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, Generate())
}

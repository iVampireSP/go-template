package schedule

import (
	"testing"
)

func TestParseCommandActionName(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		parts, err := parseCommandActionName("workspace:clean-suspended")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(parts) != 2 || parts[0] != "workspace" || parts[1] != "clean-suspended" {
			t.Fatalf("unexpected parts: %v", parts)
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		if _, err := parseCommandActionName("workspace:"); err == nil {
			t.Fatalf("expected invalid name error")
		}
	})
}

package namegen

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	suffix := Generate()

	if len(suffix) != 8 {
		t.Errorf("Generate() returned suffix of length %d, want 8", len(suffix))
	}

	if suffix != strings.ToLower(suffix) {
		t.Errorf("Generate() returned non-lowercase suffix: %s", suffix)
	}

	for _, c := range suffix {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Errorf("Generate() returned suffix with invalid character: %c in %s", c, suffix)
		}
	}
}

func TestGenerateWithPrefix(t *testing.T) {
	prefix := "x-"
	result := GenerateWithPrefix(prefix)

	if !strings.HasPrefix(result, prefix) {
		t.Errorf("GenerateWithPrefix(%s) = %s, want prefix %s", prefix, result, prefix)
	}

	if len(result) != len(prefix)+8 {
		t.Errorf("GenerateWithPrefix(%s) returned string of length %d, want %d", prefix, len(result), len(prefix)+8)
	}
}

func TestGenerateUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		suffix := Generate()
		if seen[suffix] {
			t.Errorf("Generate() produced duplicate suffix: %s", suffix)
		}
		seen[suffix] = true
	}

	if len(seen) != iterations {
		t.Errorf("Generate() produced %d unique suffixes out of %d iterations", len(seen), iterations)
	}
}

package config

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEscapeYAMLValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYAML bool // should the result be valid YAML
	}{
		{
			name:     "simple value",
			input:    "hello",
			wantYAML: true,
		},
		{
			name:     "empty value",
			input:    "",
			wantYAML: true,
		},
		{
			name:     "value with colon",
			input:    "host:port",
			wantYAML: true,
		},
		{
			name:     "multiline PEM key",
			input:    "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn\n-----END RSA PRIVATE KEY-----",
			wantYAML: true,
		},
		{
			name:     "value with quotes",
			input:    `say "hello"`,
			wantYAML: true,
		},
		{
			name:     "value with backslash",
			input:    `path\to\file`,
			wantYAML: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := escapeYAMLValue(tt.input)

			// Test that the escaped value can be parsed as YAML
			yamlContent := "key: " + escaped
			var data map[string]string
			err := yaml.Unmarshal([]byte(yamlContent), &data)
			if tt.wantYAML && err != nil {
				t.Errorf("escapeYAMLValue() produced invalid YAML: %v\nYAML content:\n%s", err, yamlContent)
				return
			}

			// Verify the parsed value matches the original input
			if tt.wantYAML && data["key"] != tt.input {
				t.Errorf("escapeYAMLValue() value mismatch\ngot:  %q\nwant: %q", data["key"], tt.input)
			}
		})
	}
}

func TestParseEnvFile_MultilineQuotedValue(t *testing.T) {
	const key = "VM_ADMIN_SSH_PRIVATE_KEY_TEST"
	t.Setenv(key, "")

	parseEnvFile(strings.Join([]string{
		key + "=\"-----BEGIN OPENSSH PRIVATE KEY-----",
		"line-1",
		"line-2",
		"-----END OPENSSH PRIVATE KEY-----\"",
	}, "\n"))

	got := os.Getenv(key)
	want := strings.Join([]string{
		"-----BEGIN OPENSSH PRIVATE KEY-----",
		"line-1",
		"line-2",
		"-----END OPENSSH PRIVATE KEY-----",
	}, "\n")
	if got != want {
		t.Fatalf("parseEnvFile() multiline value mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestParseEnvFile_SingleLineQuotedValue(t *testing.T) {
	const key = "SIMPLE_QUOTED_ENV_TEST"
	t.Setenv(key, "")

	parseEnvFile(key + "=\"hello-world\"")

	got := os.Getenv(key)
	if got != "hello-world" {
		t.Fatalf("parseEnvFile() single-line quoted value mismatch, got %q", got)
	}
}

func TestParseEnvFile_QuotedValueWithTrailingComment(t *testing.T) {
	const (
		firstKey  = "QUOTED_WITH_COMMENT_ENV_TEST"
		secondKey = "FOLLOWING_ENV_TEST"
	)
	t.Setenv(firstKey, "")
	t.Setenv(secondKey, "")

	parseEnvFile(strings.Join([]string{
		firstKey + `="hello" # inline comment`,
		secondKey + "=world",
	}, "\n"))

	if got := os.Getenv(firstKey); got != "hello" {
		t.Fatalf("parseEnvFile() quoted value with trailing comment mismatch, got %q", got)
	}
	if got := os.Getenv(secondKey); got != "world" {
		t.Fatalf("parseEnvFile() should preserve following key, got %q", got)
	}
}

func TestParseEnvFile_QuotedValueWithTrailingTextKeepsFollowingKey(t *testing.T) {
	const (
		firstKey  = "QUOTED_WITH_TEXT_ENV_TEST"
		secondKey = "FOLLOWING_ENV_TEXT_TEST"
	)
	t.Setenv(firstKey, "")
	t.Setenv(secondKey, "")

	parseEnvFile(strings.Join([]string{
		firstKey + `="hello" trailing`,
		secondKey + "=world",
	}, "\n"))

	if got := os.Getenv(firstKey); got != `"hello" trailing` {
		t.Fatalf("parseEnvFile() quoted value with trailing text mismatch, got %q", got)
	}
	if got := os.Getenv(secondKey); got != "world" {
		t.Fatalf("parseEnvFile() should preserve following key, got %q", got)
	}
}

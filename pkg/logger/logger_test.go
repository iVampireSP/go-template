package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "basic logger",
			config: Config{
				Level: "info",
				Debug: false,
			},
		},
		{
			name: "debug logger",
			config: Config{
				Level: "debug",
				Debug: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, sugar := New(tt.config)
			if logger == nil {
				t.Fatal("expected logger to be created")
			}
			if sugar == nil {
				t.Fatal("expected sugar logger to be created")
			}

			// Test logging methods using sugar logger
			sugar.Info("test info message")
			sugar.Debug("test debug message")
			sugar.Warn("test warn message")
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"invalid", "info"}, // default to info
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			level := getLogLevel(tt.level)
			if level.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, level.String())
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	logger, sugar := New(Config{
		Level: "debug",
		Debug: true,
	})

	// Create Logger wrapper for testing methods
	loggerWrapper := &Logger{
		Logger: logger,
		Sugar:  sugar,
	}

	// Test all logging methods (structured)
	loggerWrapper.Debug("debug message")
	loggerWrapper.Debug("debug structured", "key", "value")
	loggerWrapper.Info("info message")
	loggerWrapper.Info("info structured", "count", 42)
	loggerWrapper.Warn("warn message")
	loggerWrapper.Warn("warn structured", "reason", "test")
	loggerWrapper.Error("error message")
	loggerWrapper.Error("error structured", "error", "something failed")

	// Test With method
	contextLogger := loggerWrapper.With("key", "value")
	if contextLogger == nil {
		t.Error("expected With to return a logger")
	}
	contextLogger.Info("message with context")

	// Test Sync
	if err := loggerWrapper.Sync(); err != nil {
		// Sync may fail on stdout/stderr, which is acceptable
		t.Logf("Sync returned error (may be expected): %v", err)
	}
}

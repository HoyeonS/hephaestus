package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create custom config
	config := DefaultConfig().
		WithOutput(&buf).
		WithPrefix("TEST").
		WithMinLevel(INFO).
		WithTimeFormat("2006-01-02")

	// Create logger with config
	log := New(config)

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		level    string
		expected bool
	}{
		{
			name:     "info message",
			logFunc:  log.Info,
			message:  "test info",
			level:    "INFO",
			expected: true,
		},
		{
			name:     "warn message",
			logFunc:  log.Warn,
			message:  "test warn",
			level:    "WARN",
			expected: true,
		},
		{
			name:     "error message",
			logFunc:  log.Error,
			message:  "test error",
			level:    "ERROR",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			buf.Reset()

			// Log message
			tt.logFunc(tt.message)

			// Get output
			output := buf.String()

			// Check prefix
			if !strings.Contains(output, "[TEST]") {
				t.Error("log message missing prefix")
			}

			// Check level
			if !strings.Contains(output, "["+tt.level+"]") {
				t.Error("log message missing level")
			}

			// Check message
			if !strings.Contains(output, tt.message) {
				t.Error("log message missing content")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer

	// Create logger that only logs ERROR level
	config := DefaultConfig().
		WithOutput(&buf).
		WithMinLevel(ERROR)

	log := New(config)

	// Info should not be logged
	log.Info("test info")
	if buf.Len() > 0 {
		t.Error("INFO message was logged when min level is ERROR")
	}

	// Warn should not be logged
	log.Warn("test warn")
	if buf.Len() > 0 {
		t.Error("WARN message was logged when min level is ERROR")
	}

	// Error should be logged
	log.Error("test error")
	if buf.Len() == 0 {
		t.Error("ERROR message was not logged")
	}
}

func TestFormatting(t *testing.T) {
	var buf bytes.Buffer

	// Create logger with custom format
	config := DefaultConfig().
		WithOutput(&buf).
		WithoutTimestamp().
		WithPrefix("TEST")

	log := New(config)

	// Test with format string
	log.Info("count: %d, name: %s", 42, "test")

	output := buf.String()
	expected := "[TEST] [INFO] count: 42, name: test\n"

	if output != expected {
		t.Errorf("unexpected format output:\ngot:  %s\nwant: %s", output, expected)
	}
} 
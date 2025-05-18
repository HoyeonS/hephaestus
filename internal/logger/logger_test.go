package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitializeLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config with stdout",
			config: &Config{
				Level:      "info",
				OutputPath: "stdout",
			},
			expectError: false,
		},
		{
			name: "valid config with file output",
			config: &Config{
				Level:      "debug",
				OutputPath: "test.log",
			},
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				Level:      "invalid",
				OutputPath: "stdout",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger before each test
			globalLogger = nil

			err := Initialize(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, globalLogger)
			}

			// Clean up test log file if created
			if tt.config != nil && tt.config.OutputPath != "stdout" {
				os.Remove(tt.config.OutputPath)
			}
		})
	}
}

func TestLoggingWithContext(t *testing.T) {
	// Setup test config
	config := &Config{
		Level:      "debug",
		OutputPath: "stdout",
	}

	err := Initialize(config)
	assert.NoError(t, err)
	defer Sync()

	// Create context with trace ID
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-123")

	// Test all logging levels
	Debug(ctx, "debug message", zap.String("key", "value"))
	Info(ctx, "info message", zap.Int("count", 42))
	Warn(ctx, "warn message", zap.Bool("active", true))
	Error(ctx, "error message", zap.String("error", "test error"))

	// Test WithContext
	logger := WithContext(ctx)
	assert.NotNil(t, logger)
}

func TestLoggingToFile(t *testing.T) {
	// Create temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &Config{
		Level:      "info",
		OutputPath: logFile,
	}

	err := Initialize(config)
	assert.NoError(t, err)
	defer Sync()

	// Log some messages
	ctx := context.Background()
	Info(ctx, "test message")

	// Verify log file exists
	_, err = os.Stat(logFile)
	assert.NoError(t, err)
}

func TestSyncLogger(t *testing.T) {
	// Test sync with nil logger
	globalLogger = nil
	err := Sync()
	assert.NoError(t, err)

	// Test sync with initialized logger
	config := &Config{
		Level:      "info",
		OutputPath: "stdout",
	}

	err = Initialize(config)
	assert.NoError(t, err)

	err = Sync()
	assert.NoError(t, err)
}

func TestLoggingMethods(t *testing.T) {
	// Initialize logger for tests
	err := Initialize(&Config{
		Level:      "debug",
		OutputPath: "json",
	})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-id")
	fields := []zapcore.Field{
		zap.String("key", "value"),
		zap.Int("count", 42),
	}

	t.Run("debug logging", func(t *testing.T) {
		Debug(ctx, "debug message", fields...)
	})

	t.Run("info logging", func(t *testing.T) {
		Info(ctx, "info message", fields...)
	})

	t.Run("warn logging", func(t *testing.T) {
		Warn(ctx, "warn message", fields...)
	})

	t.Run("error logging", func(t *testing.T) {
		Error(ctx, "error message", fields...)
	})
}

func TestLogging(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}

	// Initialize logger with file output
	err := Initialize(&Config{
		Level:      "debug",
		OutputPath: "json",
	})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-id")

	// Test all log levels
	Debug(ctx, "debug message", zap.String("key", "value"))
	Info(ctx, "info message", zap.Int("count", 1))
	Warn(ctx, "warn message", zap.Bool("flag", true))
	Error(ctx, "error message", zap.Error(fmt.Errorf("test error")))

	// Sync to ensure all logs are written
	require.NoError(t, Sync())

	// Read log file
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	// Verify log content
	logContent := string(content)
	assert.Contains(t, logContent, "debug message")
	assert.Contains(t, logContent, "info message")
	assert.Contains(t, logContent, "warn message")
	assert.Contains(t, logContent, "error message")
	assert.Contains(t, logContent, "test-trace-id")
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context with trace ID",
			ctx:  context.WithValue(context.Background(), "trace_id", "test-trace-id"),
			want: "test-trace-id",
		},
		{
			name: "context without trace ID",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "nil context",
			ctx:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTraceID(tt.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

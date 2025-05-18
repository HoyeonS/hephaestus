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

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid debug level and json format",
			config: &Config{
				Level:  "debug",
				Format: "json",
				Output: "",
			},
			wantErr: false,
		},
		{
			name: "valid info level and console format",
			config: &Config{
				Level:  "info",
				Format: "console",
				Output: "",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				Level:  "invalid",
				Format: "json",
				Output: "",
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			_ = Sync()
		})
	}
}

func TestLoggingWithContext(t *testing.T) {
	// Setup temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Initialize logger
	config := &Config{
		Level:  "debug",
		Format: "json",
		Output: logFile,
	}
	err := Initialize(config)
	require.NoError(t, err)
	defer func() {
		_ = Sync()
	}()

	// Test with trace ID
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-id")
	Debug(ctx, "debug message", zap.String("key", "value"))
	Info(ctx, "info message", zap.Int("count", 42))
	Warn(ctx, "warn message", zap.Bool("flag", true))
	Error(ctx, "error message", zap.Error(assert.AnError))

	// Test without trace ID
	ctx = context.Background()
	Debug(ctx, "debug message without trace")
	Info(ctx, "info message without trace")
	Warn(ctx, "warn message without trace")
	Error(ctx, "error message without trace")

	// Verify log file exists and has content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestLoggingToFile(t *testing.T) {
	// Setup temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Initialize logger with file output
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: logFile,
	}
	err := Initialize(config)
	require.NoError(t, err)

	// Log some messages
	ctx := context.Background()
	Info(ctx, "test message 1")
	Error(ctx, "test message 2", zap.Error(assert.AnError))

	// Sync and verify file content
	err = Sync()
	require.NoError(t, err)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestWithContext(t *testing.T) {
	// Initialize logger
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "",
	}
	err := Initialize(config)
	require.NoError(t, err)

	// Test with trace ID
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-id")
	logger := WithContext(ctx)
	assert.NotNil(t, logger)

	// Test without trace ID
	ctx = context.Background()
	logger = WithContext(ctx)
	assert.NotNil(t, logger)

	// Test with nil context
	logger = WithContext(nil)
	assert.NotNil(t, logger)
}

func TestSync(t *testing.T) {
	// Test with initialized logger
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "",
	}
	err := Initialize(config)
	require.NoError(t, err)
	err = Sync()
	assert.NoError(t, err)

	// Test with nil logger
	globalLogger = nil
	err = Sync()
	assert.NoError(t, err)
}

func TestLoggingMethods(t *testing.T) {
	// Initialize logger for tests
	err := Initialize(Config{
		Level:  "debug",
		Format: "json",
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
	err := Initialize(Config{
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{logFile},
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
			got := GetTraceID(tt.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

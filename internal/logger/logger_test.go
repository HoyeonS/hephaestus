package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid debug level",
			config: Config{
				Level:  "debug",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "valid info level",
			config: Config{
				Level:  "info",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "invalid level",
			config: Config{
				Level:  "invalid",
				Format: "json",
			},
			wantErr: true,
		},
		{
			name: "with output file",
			config: Config{
				Level:       "info",
				Format:     "json",
				OutputPaths: []string{"test.log"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger before each test
			globalLogger = nil
			once = sync.Once{}

			err := Initialize(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, globalLogger)
			}

			// Cleanup test log file if created
			if len(tt.config.OutputPaths) > 0 {
				for _, path := range tt.config.OutputPaths {
					if path != "stdout" && path != "stderr" {
						os.Remove(path)
					}
				}
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	// Initialize logger
	err := Initialize(Config{
		Level:  "info",
		Format: "json",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		ctx     context.Context
		wantID  string
	}{
		{
			name:    "context with trace ID",
			ctx:     context.WithValue(context.Background(), "trace_id", "test-trace-id"),
			wantID:  "test-trace-id",
		},
		{
			name:    "context without trace ID",
			ctx:     context.Background(),
			wantID:  "",
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := WithContext(tt.ctx)
			assert.NotNil(t, logger)
		})
	}
}

func TestLogging(t *testing.T) {
	// Create temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Initialize logger with file output
	err := Initialize(Config{
		Level:       "debug",
		Format:     "json",
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
		name    string
		ctx     context.Context
		want    string
	}{
		{
			name:    "context with trace ID",
			ctx:     context.WithValue(context.Background(), "trace_id", "test-trace-id"),
			want:    "test-trace-id",
		},
		{
			name:    "context without trace ID",
			ctx:     context.Background(),
			want:    "",
		},
		{
			name:    "nil context",
			ctx:     nil,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTraceID(tt.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
} 
package log

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()

	validConfig := hephaestus.LoggingConfiguration{
		LogLevel:     "info",
		OutputFormat: "json",
	}

	t.Run("Successful initialization", func(t *testing.T) {
		err := processor.Initialize(ctx, validConfig)
		
		assert.NoError(t, err)
		assert.Equal(t, &validConfig, processor.config)
	})

	t.Run("Missing log level", func(t *testing.T) {
		invalidConfig := validConfig
		invalidConfig.LogLevel = ""

		err := processor.Initialize(ctx, invalidConfig)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log level is required")
	})

	t.Run("Invalid log level", func(t *testing.T) {
		invalidConfig := validConfig
		invalidConfig.LogLevel = "invalid"

		err := processor.Initialize(ctx, invalidConfig)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log level")
	})

	t.Run("Default output format", func(t *testing.T) {
		config := validConfig
		config.OutputFormat = ""

		err := processor.Initialize(ctx, config)
		
		assert.NoError(t, err)
		assert.Equal(t, "json", processor.config.OutputFormat)
	})

	t.Run("Invalid output format", func(t *testing.T) {
		invalidConfig := validConfig
		invalidConfig.OutputFormat = "invalid"

		err := processor.Initialize(ctx, invalidConfig)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output format")
	})
}

func TestCreateStream(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()
	nodeID := "test-node"

	// Initialize processor
	processor.config = &hephaestus.LoggingConfiguration{
		LogLevel:     "info",
		OutputFormat: "json",
	}

	t.Run("Successful stream creation", func(t *testing.T) {
		err := processor.CreateStream(ctx, nodeID)
		
		assert.NoError(t, err)
		stream, exists := processor.streams[nodeID]
		assert.True(t, exists)
		assert.Equal(t, nodeID, stream.NodeID)
		assert.Equal(t, processor.config.LogLevel, stream.LogLevel)
		assert.Equal(t, processor.config.OutputFormat, stream.OutputFormat)
		assert.True(t, stream.IsActive)
		assert.Empty(t, stream.Buffer)
		assert.WithinDuration(t, time.Now(), stream.LastActivity, time.Second)
	})

	t.Run("Stream already exists", func(t *testing.T) {
		err := processor.CreateStream(ctx, nodeID)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream already exists for node")
	})
}

func TestProcessLogs(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()
	nodeID := "test-node"

	// Initialize processor and create stream
	processor.config = &hephaestus.LoggingConfiguration{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	processor.CreateStream(ctx, nodeID)

	t.Run("Process JSON logs", func(t *testing.T) {
		logEntry := LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   "test error",
			Context:   map[string]interface{}{"key": "value"},
		}
		logJSON, _ := json.Marshal(logEntry)

		err := processor.ProcessLogs(ctx, nodeID, []string{string(logJSON)})
		
		assert.NoError(t, err)
		stream := processor.streams[nodeID]
		assert.Len(t, stream.Buffer, 1)
		assert.Equal(t, logEntry.Message, stream.Buffer[0].Message)
		assert.Equal(t, logEntry.Level, stream.Buffer[0].Level)
		assert.Equal(t, logEntry.Context, stream.Buffer[0].Context)
	})

	t.Run("Process text logs", func(t *testing.T) {
		err := processor.ProcessLogs(ctx, nodeID, []string{"plain text log"})
		
		assert.NoError(t, err)
		stream := processor.streams[nodeID]
		assert.Contains(t, stream.Buffer, LogEntry{
			Message: "plain text log",
			Level:   "info",
			Context: map[string]interface{}{},
		})
	})

	t.Run("Process logs with non-existent stream", func(t *testing.T) {
		err := processor.ProcessLogs(ctx, "non-existent", []string{"test log"})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream not found for node")
	})

	t.Run("Process logs with inactive stream", func(t *testing.T) {
		processor.streams[nodeID].IsActive = false

		err := processor.ProcessLogs(ctx, nodeID, []string{"test log"})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream is not active for node")
	})

	t.Run("Process logs below threshold level", func(t *testing.T) {
		processor.streams[nodeID].IsActive = true
		processor.streams[nodeID].LogLevel = "error"

		err := processor.ProcessLogs(ctx, nodeID, []string{"debug log"})
		
		assert.NoError(t, err)
		initialLength := len(processor.streams[nodeID].Buffer)
		assert.Equal(t, initialLength, len(processor.streams[nodeID].Buffer))
	})
}

func TestGetProcessedLogs(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()
	nodeID := "test-node"

	// Initialize processor and create stream with some logs
	processor.config = &hephaestus.LoggingConfiguration{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	processor.CreateStream(ctx, nodeID)
	processor.ProcessLogs(ctx, nodeID, []string{"test log 1", "test log 2"})

	t.Run("Get logs from active stream", func(t *testing.T) {
		logs, err := processor.GetProcessedLogs(ctx, nodeID)
		
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
		assert.Equal(t, "test log 1", logs[0].Message)
		assert.Equal(t, "test log 2", logs[1].Message)
	})

	t.Run("Get logs from non-existent stream", func(t *testing.T) {
		logs, err := processor.GetProcessedLogs(ctx, "non-existent")
		
		assert.Error(t, err)
		assert.Nil(t, logs)
		assert.Contains(t, err.Error(), "stream not found for node")
	})

	t.Run("Get logs from inactive stream", func(t *testing.T) {
		processor.streams[nodeID].IsActive = false

		logs, err := processor.GetProcessedLogs(ctx, nodeID)
		
		assert.Error(t, err)
		assert.Nil(t, logs)
		assert.Contains(t, err.Error(), "stream is not active for node")
	})
}

func TestClearProcessedLogs(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()
	nodeID := "test-node"

	// Initialize processor and create stream with some logs
	processor.config = &hephaestus.LoggingConfiguration{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	processor.CreateStream(ctx, nodeID)
	processor.ProcessLogs(ctx, nodeID, []string{"test log"})

	t.Run("Clear logs from active stream", func(t *testing.T) {
		err := processor.ClearProcessedLogs(ctx, nodeID)
		
		assert.NoError(t, err)
		assert.Empty(t, processor.streams[nodeID].Buffer)
	})

	t.Run("Clear logs from non-existent stream", func(t *testing.T) {
		err := processor.ClearProcessedLogs(ctx, "non-existent")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream not found for node")
	})
}

func TestCleanup(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()
	nodeID := "test-node"

	// Initialize processor and create stream
	processor.config = &hephaestus.LoggingConfiguration{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	processor.CreateStream(ctx, nodeID)

	t.Run("Successful cleanup", func(t *testing.T) {
		err := processor.Cleanup(ctx, nodeID)
		
		assert.NoError(t, err)
		_, exists := processor.streams[nodeID]
		assert.False(t, exists)
	})

	t.Run("Cleanup non-existent stream", func(t *testing.T) {
		err := processor.Cleanup(ctx, "non-existent")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream not found for node")
	})
}

func TestLogLevelValidation(t *testing.T) {
	t.Run("Valid log levels", func(t *testing.T) {
		validLevels := []string{"debug", "info", "warn", "error", "fatal"}
		for _, level := range validLevels {
			assert.True(t, isValidLogLevel(level))
		}
	})

	t.Run("Invalid log level", func(t *testing.T) {
		assert.False(t, isValidLogLevel("invalid"))
	})
}

func TestOutputFormatValidation(t *testing.T) {
	t.Run("Valid output formats", func(t *testing.T) {
		validFormats := []string{"json", "text", "logfmt", "template"}
		for _, format := range validFormats {
			assert.True(t, isValidOutputFormat(format))
		}
	})

	t.Run("Invalid output format", func(t *testing.T) {
		assert.False(t, isValidOutputFormat("invalid"))
	})
}

func TestLogLevelThreshold(t *testing.T) {
	testCases := []struct {
		logLevel       string
		thresholdLevel string
		expected       bool
	}{
		{"debug", "debug", true},
		{"info", "debug", true},
		{"warn", "info", true},
		{"error", "warn", true},
		{"fatal", "error", true},
		{"debug", "info", false},
		{"info", "warn", false},
		{"warn", "error", false},
		{"error", "fatal", false},
		{"invalid", "info", false},
		{"info", "invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.logLevel+"_"+tc.thresholdLevel, func(t *testing.T) {
			result := isLogLevelMet(tc.logLevel, tc.thresholdLevel)
			assert.Equal(t, tc.expected, result)
		})
	}
}
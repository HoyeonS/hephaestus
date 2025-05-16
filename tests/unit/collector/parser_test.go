package collector_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantMessage string
		wantErr     bool
	}{
		{
			name:        "valid log line with timestamp",
			line:        "2024-03-21 10:00:00 Test log message",
			wantMessage: "Test log message",
			wantErr:     false,
		},
		{
			name:        "log line without timestamp",
			line:        "Test log message",
			wantMessage: "Test log message",
			wantErr:     false,
		},
		{
			name:        "invalid timestamp format",
			line:        "2024/03/21 10:00:00 Test log message",
			wantMessage: "2024/03/21 10:00:00 Test log message",
			wantErr:     false,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			line:    "   \t\n",
			wantErr: true,
		},
	}

	parser, err := collector.NewParser(collector.FormatText, nil, "")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseLine(tt.line)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check message
			assert.Equal(t, tt.wantMessage, result["message"])

			// Check timestamp
			timestamp, ok := result["timestamp"].(time.Time)
			assert.True(t, ok, "timestamp should be of type time.Time")
			assert.False(t, timestamp.IsZero(), "timestamp should not be zero")
		})
	}
}

func TestParser_ParseLine_TimestampHandling(t *testing.T) {
	parser, err := collector.NewParser(collector.FormatText, nil, "")
	require.NoError(t, err)

	// Test valid timestamp parsing
	line := "2024-03-21 10:00:00 Test message"
	result, err := parser.ParseLine(line)
	require.NoError(t, err)

	timestamp, ok := result["timestamp"].(time.Time)
	require.True(t, ok, "timestamp should be of type time.Time")
	assert.Equal(t, 2024, timestamp.Year())
	assert.Equal(t, time.Month(3), timestamp.Month())
	assert.Equal(t, 21, timestamp.Day())
	assert.Equal(t, 10, timestamp.Hour())
	assert.Equal(t, 0, timestamp.Minute())
	assert.Equal(t, 0, timestamp.Second())
}

func TestParser_ParseLine_MessageExtraction(t *testing.T) {
	parser, err := collector.NewParser(collector.FormatText, nil, "")
	require.NoError(t, err)

	tests := []struct {
		name        string
		line        string
		wantMessage string
	}{
		{
			name:        "message with special characters",
			line:        "2024-03-21 10:00:00 Test: [error] {data} <warning>",
			wantMessage: "Test: [error] {data} <warning>",
		},
		{
			name:        "message with multiple spaces",
			line:        "2024-03-21 10:00:00    Multiple   Spaces   Here   ",
			wantMessage: "Multiple   Spaces   Here",
		},
		{
			name:        "message with leading/trailing spaces",
			line:        "   2024-03-21 10:00:00    Trimmed Message   ",
			wantMessage: "Trimmed Message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseLine(tt.line)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMessage, result["message"])
		})
	}
}

func TestParser_ParseLine_PatternMatching(t *testing.T) {
	patterns := []string{
		`(?P<level>ERROR|WARN|INFO)\s+(?P<message>.*)`,
		`\[(?P<component>\w+)\]\s+(?P<message>.*)`,
	}

	parser, err := collector.NewParser(collector.FormatText, patterns, "")
	require.NoError(t, err)

	tests := []struct {
		name           string
		line           string
		wantLevel     string
		wantComponent string
		wantMessage   string
	}{
		{
			name:       "error message",
			line:       "2024-03-21 10:00:00 ERROR Database connection failed",
			wantLevel: "ERROR",
			wantMessage: "Database connection failed",
		},
		{
			name:         "component message",
			line:         "2024-03-21 10:00:00 [API] Request timeout",
			wantComponent: "API",
			wantMessage:   "Request timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseLine(tt.line)
			require.NoError(t, err)

			if tt.wantLevel != "" {
				assert.Equal(t, tt.wantLevel, result["level"])
			}
			if tt.wantComponent != "" {
				assert.Equal(t, tt.wantComponent, result["component"])
			}
			assert.Contains(t, result["message"], tt.wantMessage)
		})
	}
}

func TestParser_MultiplePatterns(t *testing.T) {
	patterns := []string{
		`level=(?P<level>\w+)`,
		`severity=(?P<severity>\w+)`,
		`error="(?P<error>.*?)"`,
	}

	tests := []struct {
		name     string
		line     string
		expected map[string]interface{}
	}{
		{
			name: "matches first pattern",
			line: "level=ERROR message",
			expected: map[string]interface{}{
				"level":   "ERROR",
				"message": "message",
			},
		},
		{
			name: "matches second pattern",
			line: "severity=HIGH message",
			expected: map[string]interface{}{
				"severity": "HIGH",
				"message":  "message",
			},
		},
		{
			name: "matches third pattern",
			line: "error=\"test error\" message",
			expected: map[string]interface{}{
				"error":   "test error",
				"message": "message",
			},
		},
		{
			name: "matches multiple patterns",
			line: "level=ERROR severity=HIGH error=\"test error\"",
			expected: map[string]interface{}{
				"level":    "ERROR",
				"severity": "HIGH",
				"error":    "test error",
			},
		},
		{
			name: "no matches",
			line: "random message",
			expected: map[string]interface{}{
				"message": "random message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parser
			p, err := collector.NewParser(collector.FormatText, patterns, "")
			require.NoError(t, err)

			// Parse line
			result, err := p.ParseLine(tt.line)
			require.NoError(t, err)

			// Verify results
			for key, expectedValue := range tt.expected {
				actualValue, ok := result[key]
				require.True(t, ok, "missing key: %s", key)
				assert.Equal(t, expectedValue, actualValue)
			}
		})
	}
}

func TestParser_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Create parser with multiple patterns
	patterns := []string{
		`level=(?P<level>\w+)`,
		`severity=(?P<severity>\w+)`,
		`error="(?P<error>.*?)"`,
		`id=(?P<id>\d+)`,
		`user=(?P<user>\w+)`,
	}

	p, err := collector.NewParser(collector.FormatText, patterns, "2006-01-02 15:04:05")
	require.NoError(t, err)

	// Generate test lines
	numLines := 10000
	lines := make([]string, numLines)
	for i := 0; i < numLines; i++ {
		lines[i] = fmt.Sprintf("2024-03-21 10:00:%02d level=ERROR severity=HIGH error=\"test error %d\" id=%d user=testuser", i%60, i, i)
	}

	// Measure parsing performance
	start := time.Now()
	for _, line := range lines {
		_, err := p.ParseLine(line)
		require.NoError(t, err)
	}
	duration := time.Since(start)

	// Log performance metrics
	linesPerSecond := float64(numLines) / duration.Seconds()
	t.Logf("Parsed %d lines in %v (%.2f lines/sec)", numLines, duration, linesPerSecond)

	// Assert minimum performance requirements
	assert.Greater(t, linesPerSecond, float64(1000), "Parser should handle at least 1000 lines per second")
} 
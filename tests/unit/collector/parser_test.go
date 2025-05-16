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
	timeFormat := "2006-01-02 15:04:05"
	
	tests := []struct {
		name          string
		format        collector.LogFormat
		patterns      []string
		timeFormat    string
		line          string
		expectedError bool
		expectedTime  time.Time
		expectedLevel string
		expectedMsg   string
	}{
		{
			name:       "valid log line with timestamp",
			format:     collector.FormatText,
			patterns:   nil,
			timeFormat: timeFormat,
			line:       "2024-03-21 10:00:00 ERROR Test message",
			expectedTime: time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC),
			expectedLevel: "ERROR",
			expectedMsg:   "Test message",
		},
		{
			name:          "invalid timestamp format",
			format:        collector.FormatText,
			patterns:      nil,
			timeFormat:    timeFormat,
			line:          "invalid-time ERROR Test message",
			expectedError: true,
		},
		{
			name:          "empty line",
			format:        collector.FormatText,
			patterns:      nil,
			timeFormat:    timeFormat,
			line:          "",
			expectedError: true,
		},
		{
			name:          "missing level",
			format:        collector.FormatText,
			patterns:      nil,
			timeFormat:    timeFormat,
			line:          "2024-03-21 10:00:00 Test message",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parser
			p, err := collector.NewParser(tt.format, tt.patterns, tt.timeFormat)
			require.NoError(t, err)

			// Parse line
			result, err := p.ParseLine(tt.line)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Check timestamp
			timestamp, ok := result["timestamp"].(time.Time)
			require.True(t, ok, "timestamp should be of type time.Time")
			assert.Equal(t, tt.expectedTime, timestamp)

			// Check level
			level, ok := result["level"].(string)
			require.True(t, ok, "level should be of type string")
			assert.Equal(t, tt.expectedLevel, level)
			
			// Check message
			msg, ok := result["message"].(string)
			require.True(t, ok, "message should be of type string")
			assert.Equal(t, tt.expectedMsg, msg)
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
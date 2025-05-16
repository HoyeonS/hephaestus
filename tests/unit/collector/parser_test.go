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
		name       string
		format     collector.LogFormat
		patterns   []string
		line       string
		timeFormat string
		expected   map[string]interface{}
		wantErr    bool
	}{
		{
			name:   "json format",
			format: collector.FormatJSON,
			line:   `{"timestamp": "2024-03-21 10:00:00", "level": "ERROR", "message": "Test error"}`,
			expected: map[string]interface{}{
				"timestamp": "2024-03-21 10:00:00",
				"level":     "ERROR",
				"message":   "Test error",
			},
			wantErr: false,
		},
		{
			name:       "text format with timestamp",
			format:     collector.FormatText,
			patterns:   []string{`level=(?P<level>\w+) msg="(?P<message>.*?)"`},
			line:       "2024-03-21 10:00:00 level=ERROR msg=\"Test error\"",
			timeFormat: timeFormat,
			expected: map[string]interface{}{
				"timestamp": time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC),
				"level":     "ERROR",
				"message":   "Test error",
			},
			wantErr: false,
		},
		{
			name:   "structured format",
			format: collector.FormatStructured,
			line:   "timestamp=2024-03-21 10:00:00|level=ERROR|message=Test error",
			expected: map[string]interface{}{
				"timestamp": "2024-03-21 10:00:00",
				"level":     "ERROR",
				"message":   "Test error",
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			format:   collector.FormatJSON,
			line:     `{"invalid json`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "empty line",
			format:   collector.FormatText,
			patterns: []string{`level=(?P<level>\w+)`},
			line:     "",
			expected: map[string]interface{}{
				"message": "",
			},
			wantErr: false,
		},
		{
			name:     "no pattern match",
			format:   collector.FormatText,
			patterns: []string{`level=(?P<level>\w+)`},
			line:     "some random text",
			expected: map[string]interface{}{
				"message": "some random text",
			},
			wantErr: false,
		},
		{
			name:     "invalid pattern",
			format:   collector.FormatText,
			patterns: []string{`invalid(pattern`},
			line:     "test",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parser
			p, err := collector.NewParser(tt.format, tt.patterns, tt.timeFormat)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Parse line
			result, err := p.ParseLine(tt.line)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify results
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ExtractTimestamp(t *testing.T) {
	timeFormat := "2006-01-02 15:04:05"
	
	tests := []struct {
		name           string
		timeFormat     string
		line           string
		expectedTime   *time.Time
		expectedRemain string
	}{
		{
			name:       "valid timestamp",
			timeFormat: timeFormat,
			line:       "2024-03-21 10:00:00 ERROR Test message",
			expectedTime: func() *time.Time {
				t := time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC)
				return &t
			}(),
			expectedRemain: " ERROR Test message",
		},
		{
			name:           "no timestamp format",
			timeFormat:     "",
			line:           "ERROR Test message",
			expectedTime:   nil,
			expectedRemain: "ERROR Test message",
		},
		{
			name:           "invalid timestamp",
			timeFormat:     timeFormat,
			line:           "invalid timestamp ERROR Test message",
			expectedTime:   nil,
			expectedRemain: "invalid timestamp ERROR Test message",
		},
		{
			name:           "empty line",
			timeFormat:     timeFormat,
			line:           "",
			expectedTime:   nil,
			expectedRemain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parser
			p, err := collector.NewParser(collector.FormatText, nil, tt.timeFormat)
			require.NoError(t, err)

			// Extract timestamp
			timestamp, remain := p.ExtractTimestamp(tt.line)

			// Verify results
			if tt.expectedTime == nil {
				assert.Nil(t, timestamp)
			} else {
				assert.Equal(t, *tt.expectedTime, *timestamp)
			}
			assert.Equal(t, tt.expectedRemain, remain)
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
				"level": "ERROR",
			},
		},
		{
			name: "matches second pattern",
			line: "severity=HIGH message",
			expected: map[string]interface{}{
				"severity": "HIGH",
			},
		},
		{
			name: "matches third pattern",
			line: "error=\"test error\" message",
			expected: map[string]interface{}{
				"error": "test error",
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
			assert.Equal(t, tt.expected, result)
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
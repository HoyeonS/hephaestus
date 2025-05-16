package collector_test

import (
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetector_DetectError(t *testing.T) {
	tests := []struct {
		name        string
		patterns    map[string]collector.ErrorSeverity
		entry       map[string]interface{}
		source      string
		expectError bool
		severity    collector.ErrorSeverity
	}{
		{
			name: "basic error detection",
			patterns: map[string]collector.ErrorSeverity{
				"error":   collector.SeverityHigh,
				"failure": collector.SeverityCritical,
			},
			entry: map[string]interface{}{
				"message": "system error occurred",
				"timestamp": time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC),
			},
			source:      "test.log",
			expectError: true,
			severity:    collector.SeverityHigh,
		},
		{
			name: "critical error detection",
			patterns: map[string]collector.ErrorSeverity{
				"critical": collector.SeverityCritical,
			},
			entry: map[string]interface{}{
				"message": "critical system failure",
				"timestamp": time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC),
			},
			source:      "test.log",
			expectError: true,
			severity:    collector.SeverityCritical,
		},
		{
			name: "no error match",
			patterns: map[string]collector.ErrorSeverity{
				"error": collector.SeverityHigh,
			},
			entry: map[string]interface{}{
				"message": "normal operation",
				"timestamp": time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC),
			},
			source:      "test.log",
			expectError: false,
		},
		{
			name: "missing message field",
			patterns: map[string]collector.ErrorSeverity{
				"error": collector.SeverityHigh,
			},
			entry: map[string]interface{}{
				"timestamp": time.Date(2024, 3, 21, 10, 0, 0, 0, time.UTC),
			},
			source:      "test.log",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create detector
			d, err := collector.NewDetector(tt.patterns, 5)
			require.NoError(t, err)

			// Detect error
			result := d.DetectError(tt.entry, tt.source)

			if !tt.expectError {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.severity, result.Severity)
			assert.Equal(t, tt.source, result.Source)
			assert.NotEmpty(t, result.Pattern)
		})
	}
}

func TestDetector_AddRemovePattern(t *testing.T) {
	// Create detector
	d, err := collector.NewDetector(map[string]collector.ErrorSeverity{
		"initial": collector.SeverityHigh,
	}, 5)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name        string
		entry       map[string]interface{}
		expectError bool
		severity    collector.ErrorSeverity
	}{
		{
			name: "match initial pattern",
			entry: map[string]interface{}{
				"message": "initial error",
			},
			expectError: true,
			severity:    collector.SeverityHigh,
		},
		{
			name: "match added pattern",
			entry: map[string]interface{}{
				"message": "new error type",
			},
			expectError: true,
			severity:    collector.SeverityMedium,
		},
		{
			name: "no match after remove",
			entry: map[string]interface{}{
				"message": "initial error",
			},
			expectError: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Modify patterns based on test case
			if i == 1 {
				err := d.AddPattern("new", collector.SeverityMedium)
				require.NoError(t, err)
			} else if i == 2 {
				d.RemovePattern("initial")
			}

			// Detect error
			result := d.DetectError(tt.entry, "test.log")

			if !tt.expectError {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.severity, result.Severity)
		})
	}
}

func TestDetector_Context(t *testing.T) {
	// Create detector with context
	patterns := map[string]collector.ErrorSeverity{
		"error": collector.SeverityHigh,
	}
	d, err := collector.NewDetector(patterns, 5)
	require.NoError(t, err)

	// Test entry with context
	entry := map[string]interface{}{
		"message":   "system error occurred",
		"timestamp": time.Now(),
		"pid":       12345,
		"thread":    "main",
		"user":      "testuser",
	}

	// Detect error
	result := d.DetectError(entry, "test.log")
	require.NotNil(t, result)

	// Verify context
	assert.NotNil(t, result.Context)
	assert.Equal(t, 12345, result.Context["pid"])
	assert.Equal(t, "main", result.Context["thread"])
	assert.Equal(t, "testuser", result.Context["user"])
}

func TestDetector_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}

	// Create detector
	patterns := map[string]collector.ErrorSeverity{
		"error":     collector.SeverityHigh,
		"warning":   collector.SeverityMedium,
		"critical":  collector.SeverityCritical,
		"emergency": collector.SeverityCritical,
	}
	d, err := collector.NewDetector(patterns, 5)
	require.NoError(t, err)

	// Run concurrent detections
	numGoroutines := 10
	numDetections := 1000
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < numDetections; j++ {
				entry := map[string]interface{}{
					"message":   "concurrent error test",
					"timestamp": time.Now(),
					"routine":   routineID,
					"iteration": j,
				}
				result := d.DetectError(entry, "test.log")
				require.NotNil(t, result)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestDetector_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Create detector with multiple patterns
	patterns := make(map[string]collector.ErrorSeverity)
	for i := 0; i < 100; i++ {
		patterns[fmt.Sprintf("error%d", i)] = collector.SeverityHigh
	}

	d, err := collector.NewDetector(patterns, 5)
	require.NoError(t, err)

	// Generate test entries
	numEntries := 10000
	entries := make([]map[string]interface{}, numEntries)
	for i := 0; i < numEntries; i++ {
		entries[i] = map[string]interface{}{
			"message":   fmt.Sprintf("test error%d occurred", i%100),
			"timestamp": time.Now(),
			"id":        i,
		}
	}

	// Measure detection performance
	start := time.Now()
	for _, entry := range entries {
		result := d.DetectError(entry, "test.log")
		require.NotNil(t, result)
	}
	duration := time.Since(start)

	// Log performance metrics
	detectionsPerSecond := float64(numEntries) / duration.Seconds()
	t.Logf("Processed %d detections in %v (%.2f detections/sec)", numEntries, duration, detectionsPerSecond)

	// Assert minimum performance requirements
	assert.Greater(t, detectionsPerSecond, float64(1000), "Detector should handle at least 1000 detections per second")
} 
package collector_test

import (
	"testing"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		severity    collector.ErrorSeverity
		testMessage string
		wantErr     bool
		shouldMatch bool
	}{
		{
			name:        "add valid pattern",
			pattern:     "new_error",
			severity:    collector.SeverityHigh,
			testMessage: "new_error occurred",
			wantErr:     false,
			shouldMatch: true,
		},
		{
			name:        "add empty pattern",
			pattern:     "",
			severity:    collector.SeverityHigh,
			testMessage: "test message",
			wantErr:     true,
			shouldMatch: false,
		},
		{
			name:        "add duplicate pattern",
			pattern:     "duplicate",
			severity:    collector.SeverityHigh,
			testMessage: "duplicate error",
			wantErr:     true,
			shouldMatch: false,
		},
		{
			name:        "add pattern with invalid severity",
			pattern:     "invalid_severity",
			severity:    collector.ErrorSeverity(999),
			testMessage: "test error",
			wantErr:     true,
			shouldMatch: false,
		},
		{
			name:        "add complex pattern",
			pattern:     `error\s+\d+`,
			severity:    collector.SeverityMedium,
			testMessage: "error 123 occurred",
			wantErr:     false,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create detector with initial patterns
			initialPatterns := map[string]collector.ErrorSeverity{
				"duplicate": collector.SeverityHigh,
			}
			d, err := collector.NewDetector(initialPatterns, 5)
			require.NoError(t, err)

			// Add new pattern
			err = d.AddPattern(tt.pattern, tt.severity)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Test pattern matching
			entry := map[string]interface{}{
				"message": tt.testMessage,
			}
			result := d.DetectError(entry, "test.log")

			if tt.shouldMatch {
				require.NotNil(t, result)
				assert.Equal(t, tt.severity, result.Severity)
				assert.Contains(t, result.Pattern, tt.pattern)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestAddPattern_Concurrency(t *testing.T) {
	// Create detector
	d, err := collector.NewDetector(map[string]collector.ErrorSeverity{}, 5)
	require.NoError(t, err)

	// Add patterns concurrently
	numPatterns := 100
	errChan := make(chan error, numPatterns)

	for i := 0; i < numPatterns; i++ {
		go func(i int) {
			pattern := fmt.Sprintf("pattern%d", i)
			err := d.AddPattern(pattern, collector.SeverityHigh)
			errChan <- err
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numPatterns; i++ {
		err := <-errChan
		if err == nil {
			successCount++
		}
	}

	// Verify results
	assert.True(t, successCount > 0, "some patterns should be added successfully")
	assert.True(t, successCount < numPatterns, "some patterns should fail due to concurrent access")

	// Test pattern matching after concurrent additions
	for i := 0; i < numPatterns; i++ {
		entry := map[string]interface{}{
			"message": fmt.Sprintf("pattern%d error", i),
		}
		result := d.DetectError(entry, "test.log")
		if result != nil {
			assert.Equal(t, collector.SeverityHigh, result.Severity)
		}
	}
}

func TestAddPattern_Validation(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		severity collector.ErrorSeverity
		wantErr  bool
	}{
		{
			name:     "valid pattern and severity",
			pattern:  "error",
			severity: collector.SeverityHigh,
			wantErr:  false,
		},
		{
			name:     "pattern with special characters",
			pattern:  "error.*\\d+",
			severity: collector.SeverityMedium,
			wantErr:  false,
		},
		{
			name:     "invalid regex pattern",
			pattern:  "[invalid",
			severity: collector.SeverityHigh,
			wantErr:  true,
		},
		{
			name:     "pattern too long",
			pattern:  string(make([]byte, 1000)),
			severity: collector.SeverityHigh,
			wantErr:  true,
		},
		{
			name:     "invalid severity value",
			pattern:  "error",
			severity: collector.ErrorSeverity(-1),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create detector
			d, err := collector.NewDetector(map[string]collector.ErrorSeverity{}, 5)
			require.NoError(t, err)

			// Add pattern
			err = d.AddPattern(tt.pattern, tt.severity)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify pattern was added correctly
			entry := map[string]interface{}{
				"message": "test " + tt.pattern + " message",
			}
			result := d.DetectError(entry, "test.log")
			require.NotNil(t, result)
			assert.Equal(t, tt.severity, result.Severity)
		})
	}
} 
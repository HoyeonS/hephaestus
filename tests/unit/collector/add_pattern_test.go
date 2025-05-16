package collector_test

import (
	"fmt"
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
		wantErr     error
		shouldMatch bool
	}{
		{
			name:        "add valid pattern",
			pattern:     "new_error",
			severity:    collector.SeverityHigh,
			testMessage: "new_error occurred",
			wantErr:     nil,
			shouldMatch: true,
		},
		{
			name:        "add empty pattern",
			pattern:     "",
			severity:    collector.SeverityHigh,
			testMessage: "test message",
			wantErr:     collector.ErrEmptyPattern,
			shouldMatch: false,
		},
		{
			name:        "add duplicate pattern",
			pattern:     "duplicate",
			severity:    collector.SeverityHigh,
			testMessage: "duplicate error",
			wantErr:     collector.ErrDuplicatePattern,
			shouldMatch: false,
		},
		{
			name:        "add pattern with invalid severity",
			pattern:     "invalid_severity",
			severity:    collector.ErrorSeverity(999),
			testMessage: "test error",
			wantErr:     collector.ErrInvalidSeverity,
			shouldMatch: false,
		},
		{
			name:        "add complex pattern",
			pattern:     `error\s+\d+`,
			severity:    collector.SeverityMedium,
			testMessage: "error 123 occurred",
			wantErr:     nil,
			shouldMatch: true,
		},
		{
			name:        "add invalid regex pattern",
			pattern:     "[invalid",
			severity:    collector.SeverityHigh,
			testMessage: "test error",
			wantErr:     collector.ErrInvalidPattern,
			shouldMatch: false,
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
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			// Test pattern matching
			entry := map[string]interface{}{
				"message": tt.testMessage,
			}
			result := d.DetectError(entry, "test.log")

			if tt.shouldMatch {
				require.NotNil(t, result, "pattern should match but didn't")
				assert.Equal(t, tt.severity, result.Severity)
				assert.Contains(t, result.Pattern, tt.pattern)
			} else {
				assert.Nil(t, result, "pattern shouldn't match but did")
			}
		})
	}
}

func TestAddPattern_Concurrency(t *testing.T) {
	// Create detector with smaller capacity for better concurrency testing
	d, err := collector.NewDetector(map[string]collector.ErrorSeverity{}, 2)
	require.NoError(t, err)

	// Add patterns concurrently
	numPatterns := 10
	errChan := make(chan error, numPatterns)
	doneChan := make(chan bool, numPatterns)

	for i := 0; i < numPatterns; i++ {
		go func(i int) {
			pattern := fmt.Sprintf("pattern%d", i)
			err := d.AddPattern(pattern, collector.SeverityHigh)
			errChan <- err
			doneChan <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < numPatterns; i++ {
		<-doneChan
	}

	// Collect results
	successCount := 0
	for i := 0; i < numPatterns; i++ {
		err := <-errChan
		if err == nil {
			successCount++
		}
	}

	// Verify results - with smaller capacity, we expect more failures
	assert.True(t, successCount > 0, "some patterns should be added successfully")
	assert.True(t, successCount < numPatterns, "some patterns should fail due to capacity limits")
}

func TestAddPattern_Validation(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		severity collector.ErrorSeverity
		wantErr  error
	}{
		{
			name:     "valid pattern and severity",
			pattern:  "error",
			severity: collector.SeverityHigh,
			wantErr:  nil,
		},
		{
			name:     "pattern with special characters",
			pattern:  "error.*\\d+",
			severity: collector.SeverityMedium,
			wantErr:  nil,
		},
		{
			name:     "invalid regex pattern",
			pattern:  "[invalid",
			severity: collector.SeverityHigh,
			wantErr:  collector.ErrInvalidPattern,
		},
		{
			name:     "pattern too long",
			pattern:  string(make([]byte, 1000)),
			severity: collector.SeverityHigh,
			wantErr:  collector.ErrPatternTooLong,
		},
		{
			name:     "invalid severity value",
			pattern:  "error",
			severity: collector.ErrorSeverity(-1),
			wantErr:  collector.ErrInvalidSeverity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create detector
			d, err := collector.NewDetector(map[string]collector.ErrorSeverity{}, 5)
			require.NoError(t, err)

			// Add pattern
			err = d.AddPattern(tt.pattern, tt.severity)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			// Verify pattern was added correctly
			entry := map[string]interface{}{
				"message": fmt.Sprintf("test %s message", tt.pattern),
			}
			result := d.DetectError(entry, "test.log")
			require.NotNil(t, result, "pattern should match")
			assert.Equal(t, tt.severity, result.Severity)
		})
	}
} 
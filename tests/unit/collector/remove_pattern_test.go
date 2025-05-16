package collector_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemovePattern(t *testing.T) {
	tests := []struct {
		name           string
		initialPattern string
		removePattern  string
		testMessage    string
		shouldMatch    bool
	}{
		{
			name:           "remove existing pattern",
			initialPattern: "test_error",
			removePattern:  "test_error",
			testMessage:    "test_error occurred",
			shouldMatch:    false,
		},
		{
			name:           "remove non-existent pattern",
			initialPattern: "test_error",
			removePattern:  "nonexistent",
			testMessage:    "test_error occurred",
			shouldMatch:    true,
		},
		{
			name:           "remove empty pattern",
			initialPattern: "test_error",
			removePattern:  "",
			testMessage:    "test_error occurred",
			shouldMatch:    true,
		},
		{
			name:           "remove and verify other patterns unaffected",
			initialPattern: "error1|error2",
			removePattern:  "error1",
			testMessage:    "error2 occurred",
			shouldMatch:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create detector with initial pattern
			initialPatterns := map[string]collector.ErrorSeverity{
				tt.initialPattern: collector.SeverityHigh,
			}
			d, err := collector.NewDetector(initialPatterns, 5)
			require.NoError(t, err)

			// Remove pattern
			d.RemovePattern(tt.removePattern)

			// Test pattern matching after removal
			entry := map[string]interface{}{
				"message": tt.testMessage,
			}
			result := d.DetectError(entry, "test.log")

			if tt.shouldMatch {
				require.NotNil(t, result)
				assert.Equal(t, collector.SeverityHigh, result.Severity)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestRemovePattern_Concurrency(t *testing.T) {
	// Create detector with multiple patterns
	patterns := make(map[string]collector.ErrorSeverity)
	numPatterns := 100
	for i := 0; i < numPatterns; i++ {
		patterns[fmt.Sprintf("pattern%d", i)] = collector.SeverityHigh
	}

	d, err := collector.NewDetector(patterns, 5)
	require.NoError(t, err)

	// Remove patterns concurrently
	var wg sync.WaitGroup
	wg.Add(numPatterns)

	for i := 0; i < numPatterns; i++ {
		go func(i int) {
			defer wg.Done()
			d.RemovePattern(fmt.Sprintf("pattern%d", i))
		}(i)
	}

	// Wait for all removals to complete
	wg.Wait()

	// Verify all patterns were removed
	for i := 0; i < numPatterns; i++ {
		entry := map[string]interface{}{
			"message": fmt.Sprintf("pattern%d error", i),
		}
		result := d.DetectError(entry, "test.log")
		assert.Nil(t, result, "pattern should be removed")
	}
}

func TestRemovePattern_StateConsistency(t *testing.T) {
	// Create detector with multiple patterns
	patterns := map[string]collector.ErrorSeverity{
		"error1": collector.SeverityHigh,
		"error2": collector.SeverityMedium,
		"error3": collector.SeverityCritical,
	}

	d, err := collector.NewDetector(patterns, 5)
	require.NoError(t, err)

	// Test cases for removal and verification
	tests := []struct {
		removePattern string
		testMessage   string
		shouldMatch   bool
		severity      collector.ErrorSeverity
	}{
		{
			removePattern: "error1",
			testMessage:   "error1 occurred",
			shouldMatch:   false,
		},
		{
			removePattern: "error2",
			testMessage:   "error2 occurred",
			shouldMatch:   false,
		},
		{
			removePattern: "error3",
			testMessage:   "error3 occurred",
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		// Remove pattern
		d.RemovePattern(tt.removePattern)

		// Verify pattern was removed
		entry := map[string]interface{}{
			"message": tt.testMessage,
		}
		result := d.DetectError(entry, "test.log")
		assert.Equal(t, tt.shouldMatch, result != nil)

		// Verify other patterns still work
		for pattern, severity := range patterns {
			if pattern != tt.removePattern {
				entry := map[string]interface{}{
					"message": pattern + " occurred",
				}
				result := d.DetectError(entry, "test.log")
				if result != nil {
					assert.Equal(t, severity, result.Severity)
				}
			}
		}
	}
}

func TestRemovePattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		patterns      map[string]collector.ErrorSeverity
		removePattern string
		testMessage   string
		description   string
	}{
		{
			name: "remove pattern with special characters",
			patterns: map[string]collector.ErrorSeverity{
				`error\d+`: collector.SeverityHigh,
			},
			removePattern: `error\d+`,
			testMessage:   "error123 occurred",
			description:   "should handle regex patterns correctly",
		},
		{
			name: "remove last pattern",
			patterns: map[string]collector.ErrorSeverity{
				"last_error": collector.SeverityHigh,
			},
			removePattern: "last_error",
			testMessage:   "last_error occurred",
			description:   "should handle empty pattern list",
		},
		{
			name: "remove pattern multiple times",
			patterns: map[string]collector.ErrorSeverity{
				"duplicate": collector.SeverityHigh,
			},
			removePattern: "duplicate",
			testMessage:   "duplicate occurred",
			description:   "should handle multiple removals gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create detector
			d, err := collector.NewDetector(tt.patterns, 5)
			require.NoError(t, err)

			// Remove pattern multiple times
			for i := 0; i < 3; i++ {
				d.RemovePattern(tt.removePattern)
			}

			// Verify pattern was removed
			entry := map[string]interface{}{
				"message": tt.testMessage,
			}
			result := d.DetectError(entry, "test.log")
			assert.Nil(t, result, tt.description)
		})
	}
} 
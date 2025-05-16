package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/analyzer"
	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorAnalyzerIntegration(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "integration_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test log file
	logFile := filepath.Join(tmpDir, "test.log")
	err = os.WriteFile(logFile, []byte(""), 0644)
	require.NoError(t, err)

	// Initialize collector
	collectorConfig := collector.Config{
		LogPaths:        []string{logFile},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      10,
	}
	c, err := collector.New(collectorConfig)
	require.NoError(t, err)

	// Initialize analyzer
	analyzerConfig := analyzer.Config{
		BufferSize:      10,
		ContextWindow:   5 * time.Second,
		MinErrorSeverity: collector.SeverityMedium,
	}
	a, err := analyzer.New(analyzerConfig)
	require.NoError(t, err)

	// Start services
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop()

	err = a.Start(ctx)
	require.NoError(t, err)
	defer a.Stop()

	// Write test errors
	testCases := []struct {
		name     string
		logLines []string
		expected struct {
			errorCount   int
			minSeverity  collector.ErrorSeverity
			contextKeys  []string
			stackTraces  bool
		}
	}{
		{
			name: "basic error flow",
			logLines: []string{
				"2024-03-21 10:00:00 ERROR Test error message\n",
				"2024-03-21 10:00:01 INFO Normal operation\n",
				"2024-03-21 10:00:02 ERROR Another error\n",
			},
			expected: struct {
				errorCount   int
				minSeverity  collector.ErrorSeverity
				contextKeys  []string
				stackTraces  bool
			}{
				errorCount:  2,
				minSeverity: collector.SeverityHigh,
				contextKeys: []string{"timestamp", "message"},
				stackTraces: false,
			},
		},
		{
			name: "error with stack trace",
			logLines: []string{
				"2024-03-21 10:00:00 ERROR Exception in thread \"main\"\n",
				"java.lang.NullPointerException\n",
				"    at com.example.Class.method(Class.java:123)\n",
				"    at com.example.Main.main(Main.java:45)\n",
			},
			expected: struct {
				errorCount   int
				minSeverity  collector.ErrorSeverity
				contextKeys  []string
				stackTraces  bool
			}{
				errorCount:  1,
				minSeverity: collector.SeverityHigh,
				contextKeys: []string{"timestamp", "message", "stack_trace"},
				stackTraces: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write log lines
			f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
			require.NoError(t, err)
			for _, line := range tc.logLines {
				_, err := f.WriteString(line)
				require.NoError(t, err)
			}
			f.Close()

			// Connect collector to analyzer
			collectorChan := c.GetErrorChannel()
			analyzerChan := a.GetInputChannel()

			// Forward errors from collector to analyzer
			go func() {
				for err := range collectorChan {
					analyzerChan <- err
				}
			}()

			// Collect analyzed errors
			analyzedErrors := make([]*analyzer.AnalyzedError, 0)
			timeout := time.After(2 * time.Second)

		CollectLoop:
			for {
				select {
				case analyzed := <-a.GetOutputChannel():
					if analyzed != nil {
						analyzedErrors = append(analyzedErrors, analyzed)
					}
					if len(analyzedErrors) == tc.expected.errorCount {
						break CollectLoop
					}
				case <-timeout:
					break CollectLoop
				}
			}

			// Verify results
			assert.Equal(t, tc.expected.errorCount, len(analyzedErrors))

			for _, analyzed := range analyzedErrors {
				// Check severity
				assert.GreaterOrEqual(t, analyzed.Severity, tc.expected.minSeverity)

				// Check context
				for _, key := range tc.expected.contextKeys {
					assert.Contains(t, analyzed.Context, key)
				}

				// Check stack traces
				if tc.expected.stackTraces {
					assert.Contains(t, analyzed.Context, "stack_trace")
					assert.NotEmpty(t, analyzed.Context["stack_trace"])
				}
			}
		})
	}
}

func TestCollectorAnalyzerPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "performance_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize services
	collectorConfig := collector.Config{
		LogPaths:        []string{filepath.Join(tmpDir, "*.log")},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      1000,
	}
	c, err := collector.New(collectorConfig)
	require.NoError(t, err)

	analyzerConfig := analyzer.Config{
		BufferSize:       1000,
		ContextWindow:    5 * time.Second,
		MinErrorSeverity: collector.SeverityMedium,
	}
	a, err := analyzer.New(analyzerConfig)
	require.NoError(t, err)

	// Start services
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop()

	err = a.Start(ctx)
	require.NoError(t, err)
	defer a.Stop()

	// Create multiple log files with errors
	numFiles := 5
	errorsPerFile := 1000
	totalErrors := numFiles * errorsPerFile

	for i := 0; i < numFiles; i++ {
		logFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.log", i))
		f, err := os.Create(logFile)
		require.NoError(t, err)

		for j := 0; j < errorsPerFile; j++ {
			_, err := f.WriteString(fmt.Sprintf("2024-03-21 10:00:%02d ERROR Test error %d-%d\n", j%60, i, j))
			require.NoError(t, err)
		}
		f.Close()
	}

	// Connect collector to analyzer
	collectorChan := c.GetErrorChannel()
	analyzerChan := a.GetInputChannel()

	go func() {
		for err := range collectorChan {
			analyzerChan <- err
		}
	}()

	// Measure processing time
	start := time.Now()
	processedErrors := 0
	timeout := time.After(10 * time.Second)

ProcessLoop:
	for {
		select {
		case err := <-a.GetOutputChannel():
			if err != nil {
				processedErrors++
			}
			if processedErrors == totalErrors {
				break ProcessLoop
			}
		case <-timeout:
			break ProcessLoop
		}
	}

	duration := time.Since(start)
	errorsPerSecond := float64(processedErrors) / duration.Seconds()

	// Log performance metrics
	t.Logf("Processed %d errors in %v (%.2f errors/sec)", processedErrors, duration, errorsPerSecond)

	// Verify performance
	assert.GreaterOrEqual(t, processedErrors, totalErrors*8/10, "Should process at least 80% of errors")
	assert.GreaterOrEqual(t, errorsPerSecond, float64(500), "Should process at least 500 errors per second")
}

func TestCollectorAnalyzerResilience(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "resilience_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize services with small buffers to test overflow
	collectorConfig := collector.Config{
		LogPaths:        []string{filepath.Join(tmpDir, "test.log")},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      5,
	}
	c, err := collector.New(collectorConfig)
	require.NoError(t, err)

	analyzerConfig := analyzer.Config{
		BufferSize:       3,
		ContextWindow:    1 * time.Second,
		MinErrorSeverity: collector.SeverityMedium,
	}
	a, err := analyzer.New(analyzerConfig)
	require.NoError(t, err)

	// Start services
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop()

	err = a.Start(ctx)
	require.NoError(t, err)
	defer a.Stop()

	// Create test log file with rapid error generation
	logFile := filepath.Join(tmpDir, "test.log")
	f, err := os.Create(logFile)
	require.NoError(t, err)

	// Write many errors quickly to test overflow handling
	for i := 0; i < 20; i++ {
		_, err := f.WriteString(fmt.Sprintf("2024-03-21 10:00:%02d ERROR Test error %d\n", i%60, i))
		require.NoError(t, err)
	}
	f.Close()

	// Connect collector to analyzer with slow processing
	collectorChan := c.GetErrorChannel()
	analyzerChan := a.GetInputChannel()

	go func() {
		for err := range collectorChan {
			// Simulate slow processing
			time.Sleep(100 * time.Millisecond)
			analyzerChan <- err
		}
	}()

	// Collect results
	processedErrors := 0
	timeout := time.After(3 * time.Second)

ProcessLoop:
	for {
		select {
		case err := <-a.GetOutputChannel():
			if err != nil {
				processedErrors++
			}
		case <-timeout:
			break ProcessLoop
		}
	}

	// Verify resilience
	assert.True(t, processedErrors > 0, "Should process some errors")
	assert.True(t, processedErrors < 20, "Should not process all errors due to overflow")
} 
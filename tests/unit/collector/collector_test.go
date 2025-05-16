package collector_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/HoyeonS/hephaestus/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "collector_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	file1 := filepath.Join(tmpDir, "test1.log")
	file2 := filepath.Join(tmpDir, "test2.log")
	err = os.WriteFile(file1, []byte("test error 1\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("test error 2\n"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name          string
		paths         []string
		patterns      map[string]collector.ErrorSeverity
		expectedCount int
		wantErr      bool
	}{
		{
			name:   "single file",
			paths:  []string{file1},
			patterns: map[string]collector.ErrorSeverity{
				"error": collector.SeverityHigh,
			},
			expectedCount: 1,
			wantErr:      false,
		},
		{
			name:   "multiple files",
			paths:  []string{file1, file2},
			patterns: map[string]collector.ErrorSeverity{
				"error": collector.SeverityHigh,
			},
			expectedCount: 2,
			wantErr:      false,
		},
		{
			name:   "invalid path",
			paths:  []string{"nonexistent.log"},
			patterns: map[string]collector.ErrorSeverity{
				"error": collector.SeverityHigh,
			},
			expectedCount: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create collector
			config := collector.Config{
				LogPaths:        tt.paths,
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			}
			c, err := collector.New(config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Start collector
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err = c.Start(ctx)
			require.NoError(t, err)
			defer c.Stop()

			// Wait for errors
			errorChan := c.GetErrorChannel()
			count := 0
			timeout := time.After(1 * time.Second)

			for {
				select {
				case err := <-errorChan:
					if err != nil {
						count++
					}
				case <-timeout:
					assert.Equal(t, tt.expectedCount, count)
					return
				}
			}
		})
	}
}

func TestCollector_FileRotation(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "collector_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create initial log file
	logFile := filepath.Join(tmpDir, "app.log")
	err = os.WriteFile(logFile, []byte("initial error\n"), 0644)
	require.NoError(t, err)

	// Create collector
	config := collector.Config{
		LogPaths:        []string{logFile},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      10,
	}
	c, err := collector.New(config)
	require.NoError(t, err)

	// Start collector
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop()

	// Get error channel
	errorChan := c.GetErrorChannel()
	errorCount := 0

	// Wait for initial error
	select {
	case err := <-errorChan:
		if err != nil {
			errorCount++
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for initial error")
	}

	// Simulate log rotation
	rotatedFile := filepath.Join(tmpDir, "app.log.1")
	err = os.Rename(logFile, rotatedFile)
	require.NoError(t, err)

	// Create new log file with different content
	err = os.WriteFile(logFile, []byte("new error after rotation\n"), 0644)
	require.NoError(t, err)

	// Wait for error from new file
	timeout := time.After(2 * time.Second)
	for {
		select {
		case err := <-errorChan:
			if err != nil {
				errorCount++
				if errorCount >= 2 {
					assert.Equal(t, 2, errorCount, "should detect errors from both original and rotated files")
					return
				}
			}
		case <-timeout:
			t.Fatal("timeout waiting for error after rotation")
		}
	}
}

func TestCollector_Concurrency(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_concurrency_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple log files
	numFiles := 5
	numErrorsPerFile := 10
	expectedTotal := numFiles * numErrorsPerFile

	config := collector.Config{
		LogPaths:        []string{filepath.Join(tmpDir, "*.log")},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      expectedTotal,
	}

	// Create collector
	c, err := collector.New(config)
	require.NoError(t, err)

	// Start collector
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)

	// Create and write to multiple files concurrently
	errCh := make(chan error, numFiles)
	for i := 0; i < numFiles; i++ {
		go func(fileNum int) {
			logFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.log", fileNum))
			f, err := os.Create(logFile)
			if err != nil {
				errCh <- err
				return
			}
			defer f.Close()

			for j := 0; j < numErrorsPerFile; j++ {
				_, err := f.WriteString(fmt.Sprintf("2024-03-21 10:00:%02d ERROR Test error file%d-%d\n", j, fileNum, j))
				if err != nil {
					errCh <- err
					return
				}
			}
			errCh <- nil
		}(i)
	}

	// Wait for all writers to complete
	for i := 0; i < numFiles; i++ {
		err := <-errCh
		require.NoError(t, err)
	}

	// Collect errors
	errorChan := c.GetErrorChannel()
	errors := make([]*models.Error, 0)
	timeout := time.After(3 * time.Second)

CollectLoop:
	for {
		select {
		case err := <-errorChan:
			errors = append(errors, err)
			if len(errors) == expectedTotal {
				break CollectLoop
			}
		case <-timeout:
			break CollectLoop
		}
	}

	// Verify results
	assert.Equal(t, expectedTotal, len(errors))

	// Stop collector
	err = c.Stop()
	assert.NoError(t, err)
}

func TestCollector_BufferOverflow(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_overflow_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config with small buffer
	config := collector.Config{
		LogPaths:        []string{filepath.Join(tmpDir, "overflow.log")},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      2,
	}

	// Create collector
	c, err := collector.New(config)
	require.NoError(t, err)

	// Start collector
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)

	// Write many errors quickly
	logFile := filepath.Join(tmpDir, "overflow.log")
	f, err := os.Create(logFile)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		_, err := f.WriteString(fmt.Sprintf("2024-03-21 10:00:%02d ERROR Test error %d\n", i, i))
		require.NoError(t, err)
	}
	f.Close()

	// Collect errors
	errorChan := c.GetErrorChannel()
	errors := make([]*models.Error, 0)
	timeout := time.After(1 * time.Second)

CollectLoop:
	for {
		select {
		case err := <-errorChan:
			errors = append(errors, err)
		case <-timeout:
			break CollectLoop
		}
	}

	// Verify results - should get some errors but not all due to buffer overflow
	assert.True(t, len(errors) > 0 && len(errors) < 10)

	// Stop collector
	err = c.Stop()
	assert.NoError(t, err)
} 
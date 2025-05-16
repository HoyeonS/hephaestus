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

	// Create test files with error patterns
	file1 := filepath.Join(tmpDir, "test1.log")
	file2 := filepath.Join(tmpDir, "test2.log")
	err = os.WriteFile(file1, []byte("ERROR test error 1\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("ERROR test error 2\n"), 0644)
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
				"ERROR": collector.SeverityHigh,
			},
			expectedCount: 1,
			wantErr:      false,
		},
		{
			name:   "multiple files",
			paths:  []string{file1, file2},
			patterns: map[string]collector.ErrorSeverity{
				"ERROR": collector.SeverityHigh,
			},
			expectedCount: 2,
			wantErr:      false,
		},
		{
			name:   "invalid path",
			paths:  []string{"nonexistent.log"},
			patterns: map[string]collector.ErrorSeverity{
				"ERROR": collector.SeverityHigh,
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
						if count >= tt.expectedCount {
							assert.Equal(t, tt.expectedCount, count)
							return
						}
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
	err = os.WriteFile(logFile, []byte("ERROR initial error\n"), 0644)
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
	err = os.WriteFile(logFile, []byte("ERROR new error after rotation\n"), 0644)
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
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "collector_concurrency_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple log files
	numFiles := 5
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		files[i] = filepath.Join(tmpDir, fmt.Sprintf("test%d.log", i))
		err = os.WriteFile(files[i], []byte(fmt.Sprintf("ERROR test error %d\n", i)), 0644)
		require.NoError(t, err)
	}

	// Create collector
	config := collector.Config{
		LogPaths:        files,
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      numFiles * 10,
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

	// Wait for errors
	timeout := time.After(3 * time.Second)
	for {
		select {
		case err := <-errorChan:
			if err != nil {
				errorCount++
				if errorCount >= numFiles {
					assert.Equal(t, numFiles, errorCount, "should detect errors from all files")
					return
				}
			}
		case <-timeout:
			t.Fatal("timeout waiting for errors")
		}
	}
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
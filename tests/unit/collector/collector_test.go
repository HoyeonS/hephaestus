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
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test cases
	tests := []struct {
		name           string
		config         collector.Config
		writeLogLines  []string
		expectedErrors int
		wantErr        bool
	}{
		{
			name: "basic error detection",
			config: collector.Config{
				LogPaths:        []string{filepath.Join(tmpDir, "test.log")},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			writeLogLines: []string{
				"2024-03-21 10:00:00 INFO Normal log line",
				"2024-03-21 10:00:01 ERROR Test error message",
				"2024-03-21 10:00:02 INFO Another normal line",
			},
			expectedErrors: 1,
			wantErr:       false,
		},
		{
			name: "multiple files",
			config: collector.Config{
				LogPaths:        []string{filepath.Join(tmpDir, "*.log")},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			writeLogLines: []string{
				"2024-03-21 10:00:00 ERROR First error",
				"2024-03-21 10:00:01 ERROR Second error",
			},
			expectedErrors: 2,
			wantErr:       false,
		},
		{
			name: "invalid path",
			config: collector.Config{
				LogPaths:        []string{"/nonexistent/path/*.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			writeLogLines:  []string{},
			expectedErrors: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test log file
			if len(tt.writeLogLines) > 0 {
				logFile := filepath.Join(tmpDir, "test.log")
				err := os.WriteFile(logFile, []byte(tt.writeLogLines[0]+"\n"), 0644)
				require.NoError(t, err)
			}

			// Create collector
			c, err := collector.New(tt.config)
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

			// Write additional log lines
			if len(tt.writeLogLines) > 1 {
				logFile := filepath.Join(tmpDir, "test.log")
				f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
				require.NoError(t, err)
				for _, line := range tt.writeLogLines[1:] {
					_, err := f.WriteString(line + "\n")
					require.NoError(t, err)
				}
				f.Close()
			}

			// Collect errors
			errorChan := c.GetErrorChannel()
			errors := make([]*models.Error, 0)
			timeout := time.After(1 * time.Second)

		CollectLoop:
			for {
				select {
				case err := <-errorChan:
					errors = append(errors, err)
					if len(errors) == tt.expectedErrors {
						break CollectLoop
					}
				case <-timeout:
					break CollectLoop
				}
			}

			// Verify results
			assert.Equal(t, tt.expectedErrors, len(errors))

			// Stop collector
			err = c.Stop()
			assert.NoError(t, err)
		})
	}
}

func TestCollector_FileRotation(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_rotation_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config
	config := collector.Config{
		LogPaths:        []string{filepath.Join(tmpDir, "rotating.log")},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      10,
	}

	// Create collector
	c, err := collector.New(config)
	require.NoError(t, err)

	// Start collector
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Start(ctx)
	require.NoError(t, err)

	// Write to original log
	logFile := filepath.Join(tmpDir, "rotating.log")
	err = os.WriteFile(logFile, []byte("2024-03-21 10:00:00 ERROR Original error\n"), 0644)
	require.NoError(t, err)

	// Simulate log rotation
	err = os.Rename(logFile, logFile+".1")
	require.NoError(t, err)

	// Write to new log
	err = os.WriteFile(logFile, []byte("2024-03-21 10:00:01 ERROR New error\n"), 0644)
	require.NoError(t, err)

	// Collect errors
	errorChan := c.GetErrorChannel()
	errors := make([]*models.Error, 0)
	timeout := time.After(2 * time.Second)

CollectLoop:
	for {
		select {
		case err := <-errorChan:
			errors = append(errors, err)
			if len(errors) == 2 {
				break CollectLoop
			}
		case <-timeout:
			break CollectLoop
		}
	}

	// Verify results
	assert.Equal(t, 2, len(errors))
	assert.Contains(t, errors[0].Message, "Original error")
	assert.Contains(t, errors[1].Message, "New error")

	// Stop collector
	err = c.Stop()
	assert.NoError(t, err)
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
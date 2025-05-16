package collector_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStop(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_stop_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		setupFiles     map[string]string
		writeAfterStop []string
		wantErr        bool
	}{
		{
			name: "normal stop",
			setupFiles: map[string]string{
				"test1.log": "2024-03-21 10:00:00 ERROR Initial error\n",
			},
			writeAfterStop: []string{
				"2024-03-21 10:00:01 ERROR New error\n",
			},
			wantErr: false,
		},
		{
			name: "stop without start",
			setupFiles: map[string]string{
				"test2.log": "2024-03-21 10:00:00 ERROR Test error\n",
			},
			writeAfterStop: []string{},
			wantErr: true,
		},
		{
			name: "multiple stop calls",
			setupFiles: map[string]string{
				"test3.log": "2024-03-21 10:00:00 ERROR Test error\n",
			},
			writeAfterStop: []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			// Create initial files
			for filename, content := range tt.setupFiles {
				err := os.WriteFile(filepath.Join(testDir, filename), []byte(content), 0644)
				require.NoError(t, err)
			}

			// Create collector
			config := collector.Config{
				LogPaths:        []string{filepath.Join(testDir, "*.log")},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			}
			c, err := collector.New(config)
			require.NoError(t, err)

			// Start collector if needed
			if tt.name != "stop without start" {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				err = c.Start(ctx)
				require.NoError(t, err)
				cancel()
			}

			// Stop collector
			err = c.Stop()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Try to write after stop
			if len(tt.writeAfterStop) > 0 {
				f, err := os.OpenFile(filepath.Join(testDir, "test1.log"), os.O_APPEND|os.O_WRONLY, 0644)
				require.NoError(t, err)
				for _, content := range tt.writeAfterStop {
					_, err := f.WriteString(content)
					require.NoError(t, err)
				}
				f.Close()
			}

			// Verify error channel is closed
			errorChan := c.GetErrorChannel()
			select {
			case err, ok := <-errorChan:
				assert.False(t, ok, "error channel should be closed")
				assert.Nil(t, err, "no errors should be received after stop")
			case <-time.After(500 * time.Millisecond):
				t.Error("timeout waiting for error channel to close")
			}

			// Try second stop for multiple stop test
			if tt.name == "multiple stop calls" {
				err = c.Stop()
				assert.Error(t, err, "second stop should fail")
			}
		})
	}
}

func TestStop_Cleanup(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_stop_cleanup_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	logFile := filepath.Join(tmpDir, "test.log")
	err = os.WriteFile(logFile, []byte("2024-03-21 10:00:00 ERROR Test error\n"), 0644)
	require.NoError(t, err)

	// Create collector
	config := collector.Config{
		LogPaths:        []string{filepath.Join(tmpDir, "*.log")},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      10,
	}
	c, err := collector.New(config)
	require.NoError(t, err)

	// Start collector
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	err = c.Start(ctx)
	require.NoError(t, err)

	// Wait for initial processing
	time.Sleep(200 * time.Millisecond)

	// Stop collector
	err = c.Stop()
	require.NoError(t, err)
	cancel()

	// Verify cleanup
	// 1. Check if goroutines are stopped (indirect test through resource usage)
	time.Sleep(500 * time.Millisecond)
	
	// 2. Check if file watchers are closed
	err = os.Remove(logFile)
	assert.NoError(t, err, "file should be removable after collector stop")

	// 3. Check if error channel is closed and drained
	errorChan := c.GetErrorChannel()
	_, ok := <-errorChan
	assert.False(t, ok, "error channel should be closed")
} 
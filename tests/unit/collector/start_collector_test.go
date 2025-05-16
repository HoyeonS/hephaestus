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

func TestStart(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "collector_start_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		setupFiles     map[string]string
		config         collector.Config
		writeAfterStart []string
		expectedErrors int
		wantErr        bool
	}{
		{
			name: "start with existing file",
			setupFiles: map[string]string{
				"test1.log": "2024-03-21 10:00:00 ERROR Initial error\n",
			},
			config: collector.Config{
				LogPaths:        []string{"*.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			writeAfterStart: []string{
				"2024-03-21 10:00:01 ERROR New error\n",
			},
			expectedErrors: 2,
			wantErr:       false,
		},
		{
			name: "start with no initial errors",
			setupFiles: map[string]string{
				"test2.log": "2024-03-21 10:00:00 INFO Normal operation\n",
			},
			config: collector.Config{
				LogPaths:        []string{"*.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			writeAfterStart: []string{
				"2024-03-21 10:00:01 ERROR New error\n",
			},
			expectedErrors: 1,
			wantErr:       false,
		},
		{
			name:       "start with non-existent path",
			setupFiles: map[string]string{},
			config: collector.Config{
				LogPaths:        []string{"/nonexistent/*.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			writeAfterStart: []string{},
			expectedErrors: 0,
			wantErr:       true,
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

			// Update config with test directory
			tt.config.LogPaths = []string{filepath.Join(testDir, tt.config.LogPaths[0])}

			// Create collector
			c, err := collector.New(tt.config)
			require.NoError(t, err)

			// Start collector
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err = c.Start(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			defer c.Stop()

			// Write additional log lines
			if len(tt.writeAfterStart) > 0 {
				time.Sleep(200 * time.Millisecond) // Wait for collector to initialize
				for _, content := range tt.writeAfterStart {
					f, err := os.OpenFile(filepath.Join(testDir, "test1.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					require.NoError(t, err)
					_, err = f.WriteString(content)
					require.NoError(t, err)
					f.Close()
				}
			}

			// Collect errors
			errorChan := c.GetErrorChannel()
			errors := 0
			timeout := time.After(1 * time.Second)

		CollectLoop:
			for {
				select {
				case err := <-errorChan:
					if err != nil {
						errors++
					}
					if errors == tt.expectedErrors {
						break CollectLoop
					}
				case <-timeout:
					break CollectLoop
				}
			}

			assert.Equal(t, tt.expectedErrors, errors)
		})
	}
} 
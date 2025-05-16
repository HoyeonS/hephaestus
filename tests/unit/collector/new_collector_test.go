package collector_test

import (
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	validConfig := collector.Config{
		LogPaths:        []string{"test.log"},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      10,
	}

	tests := []struct {
		name        string
		config      collector.Config
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid config",
			config: validConfig,
		},
		{
			name: "empty log paths",
			config: collector.Config{
				LogPaths:        []string{},
				PollingInterval: validConfig.PollingInterval,
				BufferSize:      validConfig.BufferSize,
			},
			wantErr:     true,
			errContains: "log paths cannot be empty",
		},
		{
			name: "zero polling interval",
			config: collector.Config{
				LogPaths:        validConfig.LogPaths,
				PollingInterval: 0,
				BufferSize:      validConfig.BufferSize,
			},
			wantErr:     true,
			errContains: "polling interval must be greater than zero",
		},
		{
			name: "zero buffer size",
			config: collector.Config{
				LogPaths:        validConfig.LogPaths,
				PollingInterval: validConfig.PollingInterval,
				BufferSize:      0,
			},
			wantErr:     true,
			errContains: "buffer size must be greater than zero",
		},
		{
			name: "invalid log path pattern",
			config: collector.Config{
				LogPaths:        []string{"[invalid-pattern"},
				PollingInterval: validConfig.PollingInterval,
				BufferSize:      validConfig.BufferSize,
			},
			wantErr:     true,
			errContains: "invalid log path pattern",
		},
		{
			name: "negative polling interval",
			config: collector.Config{
				LogPaths:        validConfig.LogPaths,
				PollingInterval: -1 * time.Second,
				BufferSize:      validConfig.BufferSize,
			},
			wantErr:     true,
			errContains: "polling interval must be greater than zero",
		},
		{
			name: "negative buffer size",
			config: collector.Config{
				LogPaths:        validConfig.LogPaths,
				PollingInterval: validConfig.PollingInterval,
				BufferSize:      -1,
			},
			wantErr:     true,
			errContains: "buffer size must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := collector.New(tt.config)
			if tt.wantErr {
				if assert.Error(t, err) && tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, c)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
			}
		})
	}
} 
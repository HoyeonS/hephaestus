package collector_test

import (
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  collector.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: collector.Config{
				LogPaths:        []string{"test.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			wantErr: false,
		},
		{
			name: "empty log paths",
			config: collector.Config{
				LogPaths:        []string{},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			wantErr: true,
		},
		{
			name: "zero polling interval",
			config: collector.Config{
				LogPaths:        []string{"test.log"},
				PollingInterval: 0,
				BufferSize:      10,
			},
			wantErr: true,
		},
		{
			name: "zero buffer size",
			config: collector.Config{
				LogPaths:        []string{"test.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      0,
			},
			wantErr: true,
		},
		{
			name: "invalid log path pattern",
			config: collector.Config{
				LogPaths:        []string{"[invalid-pattern"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := collector.New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, c)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
			}
		})
	}
} 
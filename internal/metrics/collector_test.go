package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.nodes)
	assert.Empty(t, collector.nodes)
}

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name:    "successful initialization",
			setup:   func() { prometheus.Unregister(prometheus.NewRegistry()) },
			wantErr: false,
		},
		{
			name: "metrics already registered",
			setup: func() {
				collector := NewCollector()
				_ = collector.Initialize(context.Background())
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			collector := NewCollector()
			err := collector.Initialize(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitializeNodeMetrics(t *testing.T) {
	ctx := context.Background()
	collector := NewCollector()
	require.NoError(t, collector.Initialize(ctx))

	tests := []struct {
		name    string
		nodeID  string
		wantErr bool
	}{
		{
			name:    "initialize new node",
			nodeID:  "node1",
			wantErr: false,
		},
		{
			name:    "initialize existing node",
			nodeID:  "node1",
			wantErr: true,
		},
		{
			name:    "initialize another node",
			nodeID:  "node2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := collector.InitializeNodeMetrics(ctx, tt.nodeID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				metrics, exists := collector.nodes[tt.nodeID]
				assert.True(t, exists)
				assert.Equal(t, tt.nodeID, metrics.NodeID)
				assert.NotZero(t, metrics.CreatedAt)
				assert.NotZero(t, metrics.LastActive)
				assert.Empty(t, metrics.StatusHistory)
			}
		})
	}
}

func TestRecordNodeStatusChange(t *testing.T) {
	ctx := context.Background()
	collector := NewCollector()
	require.NoError(t, collector.Initialize(ctx))

	tests := []struct {
		name    string
		nodeID  string
		status  string
		setup   func()
		wantErr bool
	}{
		{
			name:    "record status for non-existent node",
			nodeID:  "node1",
			status:  "active",
			setup:   func() {},
			wantErr: true,
		},
		{
			name:   "record status for existing node",
			nodeID: "node2",
			status: "active",
			setup: func() {
				_ = collector.InitializeNodeMetrics(ctx, "node2")
			},
			wantErr: false,
		},
		{
			name:   "record error status",
			nodeID: "node2",
			status: "error",
			setup:  func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := collector.RecordNodeStatusChange(ctx, tt.nodeID, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				metrics := collector.nodes[tt.nodeID]
				assert.NotEmpty(t, metrics.StatusHistory)
				lastStatus := metrics.StatusHistory[len(metrics.StatusHistory)-1]
				assert.Equal(t, tt.status, lastStatus.Status)
			}
		})
	}
}

func TestRecordLogProcessing(t *testing.T) {
	ctx := context.Background()
	collector := NewCollector()
	require.NoError(t, collector.Initialize(ctx))

	tests := []struct {
		name     string
		nodeID   string
		duration time.Duration
		logCount int
		setup    func()
		wantErr  bool
	}{
		{
			name:     "record logs for non-existent node",
			nodeID:   "node1",
			duration: time.Second,
			logCount: 10,
			setup:    func() {},
			wantErr:  true,
		},
		{
			name:     "record logs for existing node",
			nodeID:   "node2",
			duration: time.Second * 2,
			logCount: 20,
			setup: func() {
				_ = collector.InitializeNodeMetrics(ctx, "node2")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := collector.RecordLogProcessing(ctx, tt.nodeID, tt.duration, tt.logCount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordModelLatency(t *testing.T) {
	ctx := context.Background()
	collector := NewCollector()
	require.NoError(t, collector.Initialize(ctx))

	tests := []struct {
		name      string
		nodeID    string
		operation string
		duration  time.Duration
		setup     func()
		wantErr   bool
	}{
		{
			name:      "record latency for non-existent node",
			nodeID:    "node1",
			operation: "predict",
			duration:  time.Second,
			setup:     func() {},
			wantErr:   true,
		},
		{
			name:      "record latency for existing node",
			nodeID:    "node2",
			operation: "predict",
			duration:  time.Second * 2,
			setup: func() {
				_ = collector.InitializeNodeMetrics(ctx, "node2")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := collector.RecordModelLatency(ctx, tt.nodeID, tt.operation, tt.duration)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordRepositoryError(t *testing.T) {
	ctx := context.Background()
	collector := NewCollector()
	require.NoError(t, collector.Initialize(ctx))

	tests := []struct {
		name      string
		nodeID    string
		operation string
		errorType string
		setup     func()
		wantErr   bool
	}{
		{
			name:      "record error for non-existent node",
			nodeID:    "node1",
			operation: "push",
			errorType: "connection_failed",
			setup:     func() {},
			wantErr:   true,
		},
		{
			name:      "record error for existing node",
			nodeID:    "node2",
			operation: "push",
			errorType: "connection_failed",
			setup: func() {
				_ = collector.InitializeNodeMetrics(ctx, "node2")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := collector.RecordRepositoryError(ctx, tt.nodeID, tt.operation, tt.errorType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCleanupNodeMetrics(t *testing.T) {
	ctx := context.Background()
	collector := NewCollector()
	require.NoError(t, collector.Initialize(ctx))

	tests := []struct {
		name    string
		nodeID  string
		setup   func()
		wantErr bool
	}{
		{
			name:    "cleanup non-existent node",
			nodeID:  "node1",
			setup:   func() {},
			wantErr: true,
		},
		{
			name:   "cleanup existing node",
			nodeID: "node2",
			setup: func() {
				_ = collector.InitializeNodeMetrics(ctx, "node2")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := collector.CleanupNodeMetrics(ctx, tt.nodeID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				_, exists := collector.nodes[tt.nodeID]
				assert.False(t, exists)
			}
		})
	}
}

func TestStatusToValue(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected float64
	}{
		{
			name:     "active status",
			status:   "active",
			expected: 1,
		},
		{
			name:     "operational status",
			status:   "operational",
			expected: 1,
		},
		{
			name:     "error status",
			status:   "error",
			expected: 2,
		},
		{
			name:     "failed status",
			status:   "failed",
			expected: 2,
		},
		{
			name:     "unknown status",
			status:   "unknown",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := statusToValue(tt.status)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestCollector(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	collector := NewCollector(registry)

	t.Run("initialize node metrics", func(t *testing.T) {
		err := collector.InitializeNodeMetrics(context.Background(), "test-node")
		assert.NoError(t, err)
	})

	t.Run("record node status change", func(t *testing.T) {
		err := collector.RecordNodeStatusChange(context.Background(), "test-node", "operational")
		assert.NoError(t, err)
	})

	t.Run("record model latency", func(t *testing.T) {
		err := collector.RecordModelLatency(context.Background(), "test-node", time.Second)
		assert.NoError(t, err)
	})

	t.Run("record repository error", func(t *testing.T) {
		err := collector.RecordRepositoryError(context.Background(), "test-node", "create_issue")
		assert.NoError(t, err)
	})

	t.Run("cleanup node metrics", func(t *testing.T) {
		err := collector.CleanupNodeMetrics(context.Background(), "test-node")
		assert.NoError(t, err)
	})
} 
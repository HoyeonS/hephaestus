package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Collector, context.Context) {
	// Create a new registry for each test
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry

	collector := NewCollector()
	ctx := context.Background()
	err := collector.Initialize(ctx)
	require.NoError(t, err)
	return collector, ctx
}

func TestInitializeNodeMetrics(t *testing.T) {
	collector, ctx := setupTest(t)

	// Test successful initialization
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	assert.NoError(t, err)

	// Test duplicate initialization
	err = collector.InitializeNodeMetrics(ctx, "test-node")
	assert.Error(t, err)
}

func TestRecordNodeStatusChange(t *testing.T) {
	collector, ctx := setupTest(t)

	// Initialize node
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	require.NoError(t, err)

	// Test successful status change
	err = collector.RecordNodeStatusChange(ctx, "test-node", "active")
	assert.NoError(t, err)

	// Verify status history
	metrics := collector.GetCurrentMetrics()
	nodeMetrics := metrics["nodes"].(map[string]interface{})["test-node"].(map[string]interface{})
	history := nodeMetrics["status_history"].([]StatusChange)
	assert.Equal(t, 1, len(history))
	assert.Equal(t, "active", history[0].Status)

	// Test non-existent node
	err = collector.RecordNodeStatusChange(ctx, "non-existent", "active")
	assert.Error(t, err)
}

func TestRecordLogProcessing(t *testing.T) {
	collector, ctx := setupTest(t)

	// Initialize node
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	require.NoError(t, err)

	// Test successful log processing recording
	err = collector.RecordLogProcessing(ctx, "test-node", time.Second, 10)
	assert.NoError(t, err)

	// Test non-existent node
	err = collector.RecordLogProcessing(ctx, "non-existent", time.Second, 10)
	assert.Error(t, err)
}

func TestRecordModelLatency(t *testing.T) {
	collector, ctx := setupTest(t)

	// Initialize node
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	require.NoError(t, err)

	// Test successful model latency recording
	err = collector.RecordModelLatency(ctx, "test-node", "inference", time.Second)
	assert.NoError(t, err)

	// Test non-existent node
	err = collector.RecordModelLatency(ctx, "non-existent", "inference", time.Second)
	assert.Error(t, err)
}

func TestRecordRepositoryError(t *testing.T) {
	collector, ctx := setupTest(t)

	// Initialize node
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	require.NoError(t, err)

	// Test successful repository error recording
	err = collector.RecordRepositoryError(ctx, "test-node", "push", "auth_failed")
	assert.NoError(t, err)

	// Test non-existent node
	err = collector.RecordRepositoryError(ctx, "non-existent", "push", "auth_failed")
	assert.Error(t, err)
}

func TestCleanupNodeMetrics(t *testing.T) {
	collector, ctx := setupTest(t)

	// Initialize node
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	require.NoError(t, err)

	// Test successful cleanup
	err = collector.CleanupNodeMetrics(ctx, "test-node")
	assert.NoError(t, err)

	// Verify node was removed
	metrics := collector.GetCurrentMetrics()
	nodeMetrics := metrics["nodes"].(map[string]interface{})
	_, exists := nodeMetrics["test-node"]
	assert.False(t, exists)

	// Test non-existent node
	err = collector.CleanupNodeMetrics(ctx, "non-existent")
	assert.Error(t, err)
}

func TestRecordOperationMetrics(t *testing.T) {
	collector, _ := setupTest(t)

	// Test recording operation metrics
	collector.RecordOperationMetrics("test_operation", time.Second, true)
	collector.RecordOperationMetrics("test_operation", time.Millisecond*500, true)
}

func TestRecordErrorMetrics(t *testing.T) {
	collector, _ := setupTest(t)

	// Test recording error metrics
	collector.RecordErrorMetrics("test_component", assert.AnError)
	collector.RecordErrorMetrics("test_component", assert.AnError)
}

func TestGetCurrentMetrics(t *testing.T) {
	collector, ctx := setupTest(t)

	// Initialize node
	err := collector.InitializeNodeMetrics(ctx, "test-node")
	require.NoError(t, err)

	// Record some metrics
	err = collector.RecordNodeStatusChange(ctx, "test-node", "active")
	require.NoError(t, err)

	// Get current metrics
	metrics := collector.GetCurrentMetrics()

	// Verify metrics structure
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "nodes")

	nodeMetrics := metrics["nodes"].(map[string]interface{})
	assert.Contains(t, nodeMetrics, "test-node")

	testNodeMetrics := nodeMetrics["test-node"].(map[string]interface{})
	assert.Contains(t, testNodeMetrics, "created_at")
	assert.Contains(t, testNodeMetrics, "last_active")
	assert.Contains(t, testNodeMetrics, "status_history")

	history := testNodeMetrics["status_history"].([]StatusChange)
	assert.Equal(t, 1, len(history))
	assert.Equal(t, "active", history[0].Status)
} 
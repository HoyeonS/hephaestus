package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the MetricsCollectionService interface
type Collector struct {
	// Node metrics
	nodes     map[string]*NodeMetrics
	nodeMutex sync.RWMutex

	// Prometheus metrics
	operationLatency    *prometheus.HistogramVec
	operationErrors     *prometheus.CounterVec
	nodeStatusGauge     *prometheus.GaugeVec
	logProcessingGauge  *prometheus.GaugeVec
	modelLatencyHist    *prometheus.HistogramVec
	repositoryErrorCount *prometheus.CounterVec
}

// NodeMetrics represents metrics for a specific node
type NodeMetrics struct {
	NodeID        string
	CreatedAt     time.Time
	LastActive    time.Time
	StatusHistory []StatusChange
}

// StatusChange represents a node status change event
type StatusChange struct {
	Status    string
	Timestamp time.Time
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		nodes: make(map[string]*NodeMetrics),
	}
}

// Initialize sets up the metrics collector
func (c *Collector) Initialize(ctx context.Context) error {
	c.operationLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "operation_latency_seconds",
			Help: "Latency of operations in seconds",
		},
		[]string{"operation"},
	)

	c.operationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "operation_errors_total",
			Help: "Total number of operation errors",
		},
		[]string{"component"},
	)

	c.nodeStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "node_status",
			Help: "Current status of nodes",
		},
		[]string{"node_id", "status"},
	)

	c.logProcessingGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "log_processing_count",
			Help: "Number of logs processed",
		},
		[]string{"node_id"},
	)

	c.modelLatencyHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "model_latency_seconds",
			Help: "Latency of model operations in seconds",
		},
		[]string{"node_id", "operation"},
	)

	c.repositoryErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "repository_errors_total",
			Help: "Total number of repository operation errors",
		},
		[]string{"node_id", "operation", "error_type"},
	)

	// Register metrics
	metrics := []prometheus.Collector{
		c.operationLatency,
		c.operationErrors,
		c.nodeStatusGauge,
		c.logProcessingGauge,
		c.modelLatencyHist,
		c.repositoryErrorCount,
	}

	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			return fmt.Errorf("failed to register metric: %v", err)
		}
	}

	return nil
}

// InitializeNodeMetrics initializes metrics for a new node
func (c *Collector) InitializeNodeMetrics(ctx context.Context, nodeID string) error {
	c.nodeMutex.Lock()
	defer c.nodeMutex.Unlock()

	if _, exists := c.nodes[nodeID]; exists {
		return fmt.Errorf("metrics already initialized for node: %s", nodeID)
	}

	c.nodes[nodeID] = &NodeMetrics{
		NodeID:        nodeID,
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
		StatusHistory: make([]StatusChange, 0),
	}

	return nil
}

// RecordOperationMetrics records metrics for an operation
func (c *Collector) RecordOperationMetrics(operationName string, duration time.Duration, successful bool) {
	c.operationLatency.WithLabelValues(operationName).Observe(duration.Seconds())
}

// RecordErrorMetrics records error metrics for a component
func (c *Collector) RecordErrorMetrics(componentName string, err error) {
	c.operationErrors.WithLabelValues(componentName).Inc()
}

// RecordNodeStatusChange records a node status change
func (c *Collector) RecordNodeStatusChange(ctx context.Context, nodeID string, status string) error {
	c.nodeMutex.Lock()
	defer c.nodeMutex.Unlock()

	node, exists := c.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	node.LastActive = time.Now()
	node.StatusHistory = append(node.StatusHistory, StatusChange{
		Status:    status,
		Timestamp: time.Now(),
	})

	c.nodeStatusGauge.WithLabelValues(nodeID, status).Set(statusToValue(status))
	return nil
}

// RecordLogProcessing records log processing metrics
func (c *Collector) RecordLogProcessing(ctx context.Context, nodeID string, duration time.Duration, logCount int) error {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	c.logProcessingGauge.WithLabelValues(nodeID).Add(float64(logCount))
	c.operationLatency.WithLabelValues("log_processing").Observe(duration.Seconds())
	return nil
}

// RecordModelLatency records model operation latency
func (c *Collector) RecordModelLatency(ctx context.Context, nodeID string, operation string, duration time.Duration) error {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	c.modelLatencyHist.WithLabelValues(nodeID, operation).Observe(duration.Seconds())
	return nil
}

// RecordRepositoryError records repository operation errors
func (c *Collector) RecordRepositoryError(ctx context.Context, nodeID string, operation string, errorType string) error {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	c.repositoryErrorCount.WithLabelValues(nodeID, operation, errorType).Inc()
	return nil
}

// CleanupNodeMetrics removes metrics for a node
func (c *Collector) CleanupNodeMetrics(ctx context.Context, nodeID string) error {
	c.nodeMutex.Lock()
	defer c.nodeMutex.Unlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	delete(c.nodes, nodeID)

	// Remove node-specific metrics
	c.nodeStatusGauge.DeleteLabelValues(nodeID, "active")
	c.nodeStatusGauge.DeleteLabelValues(nodeID, "error")
	c.logProcessingGauge.DeleteLabelValues(nodeID)

	return nil
}

// GetCurrentMetrics retrieves current system metrics
func (c *Collector) GetCurrentMetrics() map[string]interface{} {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	metrics := make(map[string]interface{})

	// Add node metrics
	nodeMetrics := make(map[string]interface{})
	for nodeID, node := range c.nodes {
		nodeMetrics[nodeID] = map[string]interface{}{
			"created_at":      node.CreatedAt,
			"last_active":     node.LastActive,
			"status_history": node.StatusHistory,
		}
	}
	metrics["nodes"] = nodeMetrics

	return metrics
}

// Helper functions

func statusToValue(status string) float64 {
	switch status {
	case "active", "operational":
		return 1
	case "error", "failed":
		return 2
	default:
		return 0
	}
} 
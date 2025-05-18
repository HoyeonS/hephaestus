package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector implements the MetricsCollectionService interface
type Collector struct {
	// Prometheus metrics
	nodeOperations   *prometheus.CounterVec
	nodeStatus       *prometheus.GaugeVec
	logProcessing    *prometheus.HistogramVec
	modelLatency     *prometheus.HistogramVec
	repositoryErrors *prometheus.CounterVec

	// Active node metrics
	nodes     map[string]*NodeMetrics
	nodeMutex sync.RWMutex
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
	Timestamp time.Time
	Status    string
}

// NewCollector creates a new instance of the metrics collector
func NewCollector() *Collector {
	collector := &Collector{
		nodes: make(map[string]*NodeMetrics),
	}

	collector.initializeMetrics()
	return collector
}

// Initialize sets up the metrics collector
func (c *Collector) Initialize(ctx context.Context) error {
	// Register metrics with Prometheus
	if err := prometheus.Register(c.nodeOperations); err != nil {
		return fmt.Errorf("failed to register node operations metric: %v", err)
	}

	if err := prometheus.Register(c.nodeStatus); err != nil {
		return fmt.Errorf("failed to register node status metric: %v", err)
	}

	if err := prometheus.Register(c.logProcessing); err != nil {
		return fmt.Errorf("failed to register log processing metric: %v", err)
	}

	if err := prometheus.Register(c.modelLatency); err != nil {
		return fmt.Errorf("failed to register model latency metric: %v", err)
	}

	if err := prometheus.Register(c.repositoryErrors); err != nil {
		return fmt.Errorf("failed to register repository errors metric: %v", err)
	}

	return nil
}

// InitializeNodeMetrics initializes metrics collection for a node
func (c *Collector) InitializeNodeMetrics(ctx context.Context, nodeID string) error {
	c.nodeMutex.Lock()
	defer c.nodeMutex.Unlock()

	if _, exists := c.nodes[nodeID]; exists {
		return fmt.Errorf("metrics already initialized for node: %s", nodeID)
	}

	metrics := &NodeMetrics{
		NodeID:        nodeID,
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
		StatusHistory: make([]StatusChange, 0),
	}

	c.nodes[nodeID] = metrics

	// Initialize node-specific metrics
	c.nodeOperations.WithLabelValues(nodeID, "initialize").Inc()
	c.nodeStatus.WithLabelValues(nodeID).Set(1) // 1 indicates active

	return nil
}

// RecordNodeStatusChange records a node status change
func (c *Collector) RecordNodeStatusChange(ctx context.Context, nodeID string, status string) error {
	c.nodeMutex.Lock()
	defer c.nodeMutex.Unlock()

	metrics, exists := c.nodes[nodeID]
	if !exists {
		return fmt.Errorf("metrics not found for node: %s", nodeID)
	}

	// Record status change
	statusChange := StatusChange{
		Timestamp: time.Now(),
		Status:    status,
	}
	metrics.StatusHistory = append(metrics.StatusHistory, statusChange)
	metrics.LastActive = statusChange.Timestamp

	// Update Prometheus metrics
	c.nodeOperations.WithLabelValues(nodeID, "status_change").Inc()
	c.nodeStatus.WithLabelValues(nodeID).Set(statusToValue(status))

	return nil
}

// RecordLogProcessing records log processing metrics
func (c *Collector) RecordLogProcessing(ctx context.Context, nodeID string, duration time.Duration, logCount int) error {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("metrics not found for node: %s", nodeID)
	}

	// Record processing duration and log count
	c.logProcessing.WithLabelValues(nodeID, "duration").Observe(duration.Seconds())
	c.nodeOperations.WithLabelValues(nodeID, "logs_processed").Add(float64(logCount))

	return nil
}

// RecordModelLatency records model interaction latency
func (c *Collector) RecordModelLatency(ctx context.Context, nodeID string, operation string, duration time.Duration) error {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("metrics not found for node: %s", nodeID)
	}

	// Record model operation latency
	c.modelLatency.WithLabelValues(nodeID, operation).Observe(duration.Seconds())
	c.nodeOperations.WithLabelValues(nodeID, "model_operation").Inc()

	return nil
}

// RecordRepositoryError records repository operation errors
func (c *Collector) RecordRepositoryError(ctx context.Context, nodeID string, operation string, errorType string) error {
	c.nodeMutex.RLock()
	defer c.nodeMutex.RUnlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("metrics not found for node: %s", nodeID)
	}

	// Record repository error
	c.repositoryErrors.WithLabelValues(nodeID, operation, errorType).Inc()
	c.nodeOperations.WithLabelValues(nodeID, "repository_error").Inc()

	return nil
}

// CleanupNodeMetrics removes metrics for a node
func (c *Collector) CleanupNodeMetrics(ctx context.Context, nodeID string) error {
	c.nodeMutex.Lock()
	defer c.nodeMutex.Unlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return fmt.Errorf("metrics not found for node: %s", nodeID)
	}

	// Record cleanup operation
	c.nodeOperations.WithLabelValues(nodeID, "cleanup").Inc()
	c.nodeStatus.DeleteLabelValues(nodeID)

	delete(c.nodes, nodeID)
	return nil
}

// Helper functions

func (c *Collector) initializeMetrics() {
	c.nodeOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hephaestus_node_operations_total",
			Help: "Total number of node operations by type",
		},
		[]string{"node_id", "operation"},
	)

	c.nodeStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hephaestus_node_status",
			Help: "Current node status (0: inactive, 1: active, 2: error)",
		},
		[]string{"node_id"},
	)

	c.logProcessing = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hephaestus_log_processing_duration_seconds",
			Help:    "Log processing duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"node_id", "metric"},
	)

	c.modelLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hephaestus_model_latency_seconds",
			Help:    "Model operation latency in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{"node_id", "operation"},
	)

	c.repositoryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hephaestus_repository_errors_total",
			Help: "Total number of repository errors by type",
		},
		[]string{"node_id", "operation", "error_type"},
	)
}

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
package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"go.uber.org/zap"
)

// Manager implements the NodeLifecycleManager interface
type Manager struct {
	// Dependencies
	logProcessor      hephaestus.LogProcessingService
	modelService      hephaestus.ModelServiceProvider
	remoteRepoService hephaestus.RemoteRepositoryService
	metricsCollector  hephaestus.MetricsCollectionService

	// Node registry
	nodes     map[string]*hephaestus.SystemNode
	nodesMutex sync.RWMutex
}

// NewManager creates a new instance of the node manager
func NewManager(
	logProcessor hephaestus.LogProcessingService,
	modelService hephaestus.ModelServiceProvider,
	remoteRepoService hephaestus.RemoteRepositoryService,
	metricsCollector hephaestus.MetricsCollectionService,
) *Manager {
	return &Manager{
		logProcessor:      logProcessor,
		modelService:      modelService,
		remoteRepoService: remoteRepoService,
		metricsCollector:  metricsCollector,
		nodes:            make(map[string]*hephaestus.SystemNode),
	}
}

// CreateSystemNode creates a new node with the provided configuration
func (m *Manager) CreateSystemNode(ctx context.Context, config *hephaestus.SystemConfiguration) (*hephaestus.SystemNode, error) {
	if config == nil {
		logger.Error(ctx, "configuration is required")
		return nil, fmt.Errorf("configuration is required")
	}

	// Validate configuration
	if err := m.validateConfiguration(config); err != nil {
		logger.Error(ctx, "invalid configuration", zap.Error(err))
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	nodeID := uuid.New().String()
	logger.Info(ctx, "creating new node", zap.String("node_id", nodeID))

	// Initialize remote repository connection
	if err := m.remoteRepoService.Initialize(ctx, config.RemoteSettings); err != nil {
		logger.Error(ctx, "failed to initialize remote repository",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to initialize remote repository: %v", err)
	}

	// Initialize model service
	if err := m.modelService.Initialize(ctx, config.ModelSettings); err != nil {
		logger.Error(ctx, "failed to initialize model service",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to initialize model service: %v", err)
	}

	// Create new node
	node := &hephaestus.SystemNode{
		NodeIdentifier: nodeID,
		NodeConfig:     config,
		CurrentStatus:  hephaestus.NodeStatusInitializing,
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
	}

	// Store node in registry
	m.nodesMutex.Lock()
	m.nodes[node.NodeIdentifier] = node
	m.nodesMutex.Unlock()

	// Initialize metrics collection
	if err := m.metricsCollector.InitializeNodeMetrics(ctx, node.NodeIdentifier); err != nil {
		logger.Error(ctx, "failed to initialize metrics collection",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to initialize metrics collection: %v", err)
	}

	// Update node status to operational
	node.CurrentStatus = hephaestus.NodeStatusOperational
	logger.Info(ctx, "node created successfully",
		zap.String("node_id", nodeID),
		zap.String("status", string(node.CurrentStatus)),
	)

	return node, nil
}

// GetSystemNode retrieves a node by its identifier
func (m *Manager) GetSystemNode(ctx context.Context, nodeID string) (*hephaestus.SystemNode, error) {
	m.nodesMutex.RLock()
	node, exists := m.nodes[nodeID]
	m.nodesMutex.RUnlock()

	if !exists {
		logger.Warn(ctx, "node not found", zap.String("node_id", nodeID))
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	logger.Debug(ctx, "retrieved node",
		zap.String("node_id", nodeID),
		zap.String("status", string(node.CurrentStatus)),
	)
	return node, nil
}

// UpdateNodeOperationalStatus updates the operational status of a node
func (m *Manager) UpdateNodeOperationalStatus(ctx context.Context, nodeID string, status hephaestus.NodeStatus) error {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		logger.Warn(ctx, "node not found", zap.String("node_id", nodeID))
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Update status
	oldStatus := node.CurrentStatus
	node.CurrentStatus = status
	node.LastActive = time.Now()

	logger.Info(ctx, "node status updated",
		zap.String("node_id", nodeID),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(status)),
	)

	// Record status change metric
	if err := m.metricsCollector.RecordNodeStatusChange(ctx, nodeID, string(status)); err != nil {
		logger.Error(ctx, "failed to record status change metric",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to record status change metric: %v", err)
	}

	return nil
}

// DeleteSystemNode removes a node from the system
func (m *Manager) DeleteSystemNode(ctx context.Context, nodeID string) error {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		logger.Warn(ctx, "node not found", zap.String("node_id", nodeID))
		return fmt.Errorf("node not found: %s", nodeID)
	}

	logger.Info(ctx, "deleting node", zap.String("node_id", nodeID))

	// Clean up resources
	if err := m.cleanupNodeResources(ctx, node); err != nil {
		logger.Error(ctx, "failed to cleanup node resources",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to cleanup node resources: %v", err)
	}

	// Remove node from registry
	delete(m.nodes, nodeID)
	logger.Info(ctx, "node deleted successfully", zap.String("node_id", nodeID))

	return nil
}

// Helper functions

func (m *Manager) validateConfiguration(config *hephaestus.SystemConfiguration) error {
	if config.RemoteSettings.AuthToken == "" {
		return fmt.Errorf("remote repository auth token is required")
	}

	if config.ModelSettings.ServiceAPIKey == "" {
		return fmt.Errorf("model service API key is required")
	}

	if config.LoggingSettings.LogLevel == "" {
		return fmt.Errorf("log level is required")
	}

	if config.OperationalMode == "" {
		return fmt.Errorf("operational mode is required")
	}

	return nil
}

func (m *Manager) cleanupNodeResources(ctx context.Context, node *hephaestus.SystemNode) error {
	// Clean up remote repository resources
	if err := m.remoteRepoService.Cleanup(ctx, node.NodeIdentifier); err != nil {
		logger.Error(ctx, "failed to cleanup remote repository resources",
			zap.String("node_id", node.NodeIdentifier),
			zap.Error(err),
		)
		return fmt.Errorf("failed to cleanup remote repository resources: %v", err)
	}

	// Clean up model service resources
	if err := m.modelService.Cleanup(ctx, node.NodeIdentifier); err != nil {
		logger.Error(ctx, "failed to cleanup model service resources",
			zap.String("node_id", node.NodeIdentifier),
			zap.Error(err),
		)
		return fmt.Errorf("failed to cleanup model service resources: %v", err)
	}

	// Clean up metrics collection
	if err := m.metricsCollector.CleanupNodeMetrics(ctx, node.NodeIdentifier); err != nil {
		logger.Error(ctx, "failed to cleanup metrics collection",
			zap.String("node_id", node.NodeIdentifier),
			zap.Error(err),
		)
		return fmt.Errorf("failed to cleanup metrics collection: %v", err)
	}

	return nil
} 
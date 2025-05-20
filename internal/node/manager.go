package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// Manager handles node registration and status tracking
type NodeManager struct {
	nodes             map[string]*hephaestus.SystemNode
	mu                sync.RWMutex
	metricsCollector  hephaestus.MetricsCollectionService
	remoteRepoService hephaestus.RepositoryManager
}

// NewManager creates a new node manager instance
func NewManager(metrics hephaestus.MetricsCollectionService, repo hephaestus.RepositoryManager) *NodeManager {
	return &NodeManager{
		nodes:             make(map[string]*hephaestus.SystemNode),
		metricsCollector:  metrics,
		remoteRepoService: repo,
	}
}

// Initialize sets up the node manager
func (m *NodeManager) Initialize(ctx context.Context) error {
	logger.Info(ctx, "Initializing node manager")

	if err := m.remoteRepoService.Initialize(ctx); err != nil {
		logger.Error(ctx, "Failed to initialize remote repository service", logger.Field("error", err))
		return fmt.Errorf("failed to initialize remote repository service: %v", err)
	}

	return nil
}

// RegisterNode registers a new node with the system
func (m *Manager) RegisterNode(ctx context.Context, node *hephaestus.SystemNode) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[node.NodeID]; exists {
		return fmt.Errorf("node %s already registered", node.NodeID)
	}

	node.CreatedAt = time.Now()
	node.LastActive = time.Now()
	node.Status = hephaestus.NodeStatusActive

	m.nodes[node.NodeID] = node
	m.metricsCollector.RecordNodeStatus(node.NodeID, string(node.Status))

	logger.Info(ctx, "Node registered successfully",
		logger.Field("node_id", node.NodeID),
		logger.Field("status", string(node.Status)),
	)

	return nil
}

// UpdateNodeStatus updates the status of a registered node
func (m *Manager) UpdateNodeStatus(ctx context.Context, nodeID string, status hephaestus.NodeStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.Status = status
	node.LastActive = time.Now()
	m.metricsCollector.RecordNodeStatus(nodeID, string(status))

	logger.Info(ctx, "Node status updated",
		logger.Field("node_id", nodeID),
		logger.Field("status", string(status)),
	)

	return nil
}

// GetNode retrieves information about a registered node
func (m *Manager) GetNode(ctx context.Context, nodeID string) (*hephaestus.SystemNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	return node, nil
}

// ListNodes returns a list of all registered nodes
func (m *Manager) ListNodes(ctx context.Context) ([]*hephaestus.SystemNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]*hephaestus.SystemNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// RemoveNode removes a node from the system
func (m *Manager) RemoveNode(ctx context.Context, nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	delete(m.nodes, nodeID)
	m.metricsCollector.RecordNodeStatus(nodeID, string(hephaestus.NodeStatusRemoved))

	logger.Info(ctx, "Node removed successfully",
		logger.Field("node_id", nodeID),
	)

	return nil
}

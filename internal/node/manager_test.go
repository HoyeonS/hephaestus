package node

import (
	"context"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations
type MockLogProcessor struct {
	mock.Mock
}

func (m *MockLogProcessor) Initialize(ctx context.Context, config *hephaestus.LoggingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

type MockModelService struct {
	mock.Mock
}

func (m *MockModelService) Initialize(ctx context.Context, config *hephaestus.ModelConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockModelService) Cleanup(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

type MockRemoteRepo struct {
	mock.Mock
}

func (m *MockRemoteRepo) Initialize(ctx context.Context, config *hephaestus.RepositoryConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockRemoteRepo) Cleanup(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) InitializeNodeMetrics(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockMetricsCollector) RecordNodeStatusChange(ctx context.Context, nodeID string, status string) error {
	args := m.Called(ctx, nodeID, status)
	return args.Error(0)
}

func (m *MockMetricsCollector) CleanupNodeMetrics(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func setupTest(t *testing.T) (context.Context, *Manager, *MockLogProcessor, *MockModelService, *MockRemoteRepo, *MockMetricsCollector) {
	// Initialize logger for tests
	err := logger.Initialize(&logger.Config{
		Level:      "debug",
		OutputPath: "console",
	})
	require.NoError(t, err)

	ctx := context.Background()
	mockLogProcessor := new(MockLogProcessor)
	mockModelService := new(MockModelService)
	mockRemoteRepo := new(MockRemoteRepo)
	mockMetricsCollector := new(MockMetricsCollector)

	manager := NewManager(
		mockLogProcessor,
		mockModelService,
		mockRemoteRepo,
		mockMetricsCollector,
	)

	return ctx, manager, mockLogProcessor, mockModelService, mockRemoteRepo, mockMetricsCollector
}

func TestCreateSystemNode(t *testing.T) {
	ctx, manager, _, mockModelService, mockRemoteRepo, mockMetricsCollector := setupTest(t)

	t.Run("successful node creation", func(t *testing.T) {
		config := &hephaestus.SystemConfiguration{
			RemoteSettings: &hephaestus.RemoteSettings{
				AuthToken: "test-token",
			},
			ModelSettings: &hephaestus.ModelSettings{
				ServiceAPIKey: "test-key",
			},
			LoggingSettings: &hephaestus.LoggingConfiguration{
				LogLevel: "info",
			},
			OperationalMode: "active",
		}

		mockRemoteRepo.On("Initialize", ctx, config.RemoteSettings).Return(nil)
		mockModelService.On("Initialize", ctx, config.ModelSettings).Return(nil)
		mockMetricsCollector.On("InitializeNodeMetrics", ctx, mock.AnythingOfType("string")).Return(nil)

		node, err := manager.CreateSystemNode(ctx, config)
		require.NoError(t, err)
		assert.NotNil(t, node)
		assert.NotEmpty(t, node.NodeIdentifier)
		assert.Equal(t, hephaestus.NodeStatusOperational, node.CurrentStatus)

		mockRemoteRepo.AssertExpectations(t)
		mockModelService.AssertExpectations(t)
		mockMetricsCollector.AssertExpectations(t)
	})

	t.Run("nil configuration", func(t *testing.T) {
		node, err := manager.CreateSystemNode(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "configuration is required")
	})

	t.Run("invalid configuration", func(t *testing.T) {
		config := &hephaestus.SystemConfiguration{}
		node, err := manager.CreateSystemNode(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "invalid configuration")
	})
}

func TestGetSystemNode(t *testing.T) {
	ctx, manager, _, _, _, _ := setupTest(t)

	t.Run("get existing node", func(t *testing.T) {
		// Create a test node
		node := &hephaestus.SystemNode{
			NodeIdentifier: "test-node",
			CurrentStatus:  hephaestus.NodeStatusOperational,
			CreatedAt:      time.Now(),
			LastActive:     time.Now(),
		}
		manager.nodes[node.NodeIdentifier] = node

		// Retrieve the node
		retrieved, err := manager.GetSystemNode(ctx, node.NodeIdentifier)
		require.NoError(t, err)
		assert.Equal(t, node, retrieved)
	})

	t.Run("get non-existent node", func(t *testing.T) {
		node, err := manager.GetSystemNode(ctx, "non-existent")
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "node not found")
	})
}

func TestUpdateNodeOperationalStatus(t *testing.T) {
	ctx, manager, _, _, _, mockMetricsCollector := setupTest(t)

	t.Run("update existing node status", func(t *testing.T) {
		// Create a test node
		node := &hephaestus.SystemNode{
			NodeIdentifier: "test-node",
			CurrentStatus:  hephaestus.NodeStatusOperational,
			CreatedAt:      time.Now(),
			LastActive:     time.Now(),
		}
		manager.nodes[node.NodeIdentifier] = node

		mockMetricsCollector.On("RecordNodeStatusChange", ctx, node.NodeIdentifier, string(hephaestus.NodeStatusPaused)).Return(nil)

		err := manager.UpdateNodeOperationalStatus(ctx, node.NodeIdentifier, hephaestus.NodeStatusPaused)
		require.NoError(t, err)
		assert.Equal(t, hephaestus.NodeStatusPaused, node.CurrentStatus)

		mockMetricsCollector.AssertExpectations(t)
	})

	t.Run("update non-existent node status", func(t *testing.T) {
		err := manager.UpdateNodeOperationalStatus(ctx, "non-existent", hephaestus.NodeStatusPaused)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})
}

func TestDeleteSystemNode(t *testing.T) {
	ctx, manager, _, mockModelService, mockRemoteRepo, mockMetricsCollector := setupTest(t)

	t.Run("delete existing node", func(t *testing.T) {
		// Create a test node
		node := &hephaestus.SystemNode{
			NodeIdentifier: "test-node",
			CurrentStatus:  hephaestus.NodeStatusOperational,
			CreatedAt:      time.Now(),
			LastActive:     time.Now(),
		}
		manager.nodes[node.NodeIdentifier] = node

		mockRemoteRepo.On("Cleanup", ctx, node.NodeIdentifier).Return(nil)
		mockModelService.On("Cleanup", ctx, node.NodeIdentifier).Return(nil)
		mockMetricsCollector.On("CleanupNodeMetrics", ctx, node.NodeIdentifier).Return(nil)

		err := manager.DeleteSystemNode(ctx, node.NodeIdentifier)
		require.NoError(t, err)
		assert.NotContains(t, manager.nodes, node.NodeIdentifier)

		mockRemoteRepo.AssertExpectations(t)
		mockModelService.AssertExpectations(t)
		mockMetricsCollector.AssertExpectations(t)
	})

	t.Run("delete non-existent node", func(t *testing.T) {
		err := manager.DeleteSystemNode(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})
}

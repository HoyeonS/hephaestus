package model

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

type MockModelClient struct {
	mock.Mock
}

func (m *MockModelClient) GenerateSolution(ctx context.Context, problem *hephaestus.Problem, config *hephaestus.ModelSettings) (*hephaestus.Solution, error) {
	args := m.Called(ctx, problem, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*hephaestus.Solution), args.Error(1)
}

type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordModelLatency(ctx context.Context, nodeID string, duration time.Duration) error {
	args := m.Called(ctx, nodeID, duration)
	return args.Error(0)
}

func setupTest(t *testing.T) (context.Context, *Service, *MockModelClient, *MockMetricsCollector) {
	// Initialize logger for tests
	err := logger.Initialize(&logger.Config{
		Level:  "debug",
		Format: "console",
	})
	require.NoError(t, err)

	ctx := context.Background()
	mockClient := new(MockModelClient)
	mockMetrics := new(MockMetricsCollector)
	service := NewService(mockClient, mockMetrics)

	return ctx, service, mockClient, mockMetrics
}

func TestInitialize(t *testing.T) {
	ctx, service, _, _ := setupTest(t)

	t.Run("successful initialization", func(t *testing.T) {
		config := &hephaestus.ModelSettings{
			ServiceProvider: "test-provider",
			ServiceAPIKey:   "test-key",
			ModelName:      "test-model",
		}

		err := service.Initialize(ctx, config)
		require.NoError(t, err)
		assert.Equal(t, config, service.config)
	})

	t.Run("nil configuration", func(t *testing.T) {
		err := service.Initialize(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model configuration is required")
	})

	t.Run("missing API key", func(t *testing.T) {
		config := &hephaestus.ModelSettings{
			ServiceProvider: "test-provider",
			ModelName:      "test-model",
		}

		err := service.Initialize(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model service API key is required")
	})
}

func TestGenerateSolution(t *testing.T) {
	ctx, service, mockClient, mockMetrics := setupTest(t)

	config := &hephaestus.ModelSettings{
		ServiceProvider: "test-provider",
		ServiceAPIKey:   "test-key",
		ModelName:      "test-model",
	}
	err := service.Initialize(ctx, config)
	require.NoError(t, err)

	t.Run("successful solution generation", func(t *testing.T) {
		nodeID := "test-node"
		problem := &hephaestus.Problem{
			Type:        "test-type",
			Description: "test description",
			Context:     []string{"test context"},
		}
		expectedSolution := &hephaestus.Solution{
			ProblemType: "test-type",
			Solution:    "test solution",
		}

		mockClient.On("GenerateSolution", ctx, problem, config).Return(expectedSolution, nil)
		mockMetrics.On("RecordModelLatency", ctx, nodeID, mock.AnythingOfType("time.Duration")).Return(nil)

		solution, err := service.GenerateSolution(ctx, nodeID, problem)
		require.NoError(t, err)
		assert.Equal(t, expectedSolution, solution)

		mockClient.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)

		// Verify session was created
		session, exists := service.sessions[nodeID]
		assert.True(t, exists)
		assert.Equal(t, nodeID, session.NodeID)
		assert.True(t, session.IsActive)
		assert.Equal(t, config, session.Configuration)
	})

	t.Run("failed solution generation", func(t *testing.T) {
		nodeID := "test-node"
		problem := &hephaestus.Problem{
			Type:        "test-type",
			Description: "test description",
			Context:     []string{"test context"},
		}

		mockClient.On("GenerateSolution", ctx, problem, config).Return(nil, assert.AnError)

		solution, err := service.GenerateSolution(ctx, nodeID, problem)
		assert.Error(t, err)
		assert.Nil(t, solution)
		assert.Contains(t, err.Error(), "failed to generate solution")

		mockClient.AssertExpectations(t)
	})

	t.Run("failed metrics recording", func(t *testing.T) {
		nodeID := "test-node"
		problem := &hephaestus.Problem{
			Type:        "test-type",
			Description: "test description",
			Context:     []string{"test context"},
		}
		expectedSolution := &hephaestus.Solution{
			ProblemType: "test-type",
			Solution:    "test solution",
		}

		mockClient.On("GenerateSolution", ctx, problem, config).Return(expectedSolution, nil)
		mockMetrics.On("RecordModelLatency", ctx, nodeID, mock.AnythingOfType("time.Duration")).Return(assert.AnError)

		solution, err := service.GenerateSolution(ctx, nodeID, problem)
		require.NoError(t, err)
		assert.Equal(t, expectedSolution, solution)

		mockClient.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

func TestCleanup(t *testing.T) {
	ctx, service, _, _ := setupTest(t)

	config := &hephaestus.ModelSettings{
		ServiceProvider: "test-provider",
		ServiceAPIKey:   "test-key",
		ModelName:      "test-model",
	}
	err := service.Initialize(ctx, config)
	require.NoError(t, err)

	t.Run("cleanup existing session", func(t *testing.T) {
		nodeID := "test-node"
		session := &ModelSession{
			NodeID:        nodeID,
			LastActive:    time.Now(),
			IsActive:      true,
			Configuration: config,
		}
		service.sessions[nodeID] = session

		err := service.Cleanup(ctx, nodeID)
		require.NoError(t, err)
		assert.NotContains(t, service.sessions, nodeID)
	})

	t.Run("cleanup non-existent session", func(t *testing.T) {
		err := service.Cleanup(ctx, "non-existent")
		require.NoError(t, err)
	})
}

func TestCreateSession(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	nodeID := "test-node"

	// Initialize service
	service.config = &hephaestus.ModelServiceConfiguration{
		ServiceProvider: "openai",
		ModelVersion:    "gpt-4",
	}

	t.Run("Successful session creation", func(t *testing.T) {
		err := service.CreateSession(ctx, nodeID)
		
		assert.NoError(t, err)
		session, exists := service.sessions[nodeID]
		assert.True(t, exists)
		assert.Equal(t, nodeID, session.NodeID)
		assert.Equal(t, service.config.ServiceProvider, session.ServiceProvider)
		assert.Equal(t, service.config.ModelVersion, session.ModelVersion)
		assert.True(t, session.IsActive)
		assert.Empty(t, session.Context)
	})

	t.Run("Session already exists", func(t *testing.T) {
		err := service.CreateSession(ctx, nodeID)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session already exists for node")
	})
}

func TestGetSession(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	nodeID := "test-node"

	t.Run("Get existing active session", func(t *testing.T) {
		service.sessions[nodeID] = &ModelSession{
			NodeID:          nodeID,
			ServiceProvider: "openai",
			ModelVersion:    "gpt-4",
			IsActive:        true,
		}

		session, err := service.GetSession(ctx, nodeID)
		
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, nodeID, session.NodeID)
		assert.True(t, session.IsActive)
	})

	t.Run("Get non-existent session", func(t *testing.T) {
		session, err := service.GetSession(ctx, "non-existent")
		
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "session not found for node")
	})

	t.Run("Get inactive session", func(t *testing.T) {
		service.sessions[nodeID].IsActive = false

		session, err := service.GetSession(ctx, nodeID)
		
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "session is not active for node")
	})
}

func TestValidateModelVersion(t *testing.T) {
	service := NewService()

	t.Run("Valid OpenAI models", func(t *testing.T) {
		validModels := []string{"gpt-4", "gpt-4-32k", "gpt-3.5-turbo"}
		for _, model := range validModels {
			err := service.validateModelVersion("openai", model)
			assert.NoError(t, err)
		}
	})

	t.Run("Invalid OpenAI model", func(t *testing.T) {
		err := service.validateModelVersion("openai", "invalid-model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported OpenAI model version")
	})

	t.Run("Valid Anthropic models", func(t *testing.T) {
		validModels := []string{"claude-2", "claude-instant"}
		for _, model := range validModels {
			err := service.validateModelVersion("anthropic", model)
			assert.NoError(t, err)
		}
	})

	t.Run("Invalid Anthropic model", func(t *testing.T) {
		err := service.validateModelVersion("anthropic", "invalid-model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported Anthropic model version")
	})

	t.Run("Unsupported provider", func(t *testing.T) {
		err := service.validateModelVersion("invalid-provider", "model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported model provider")
	})
} 
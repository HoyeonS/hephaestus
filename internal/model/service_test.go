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

// Mock types
type MockModelClient struct {
	mock.Mock
}

func (m *MockModelClient) GenerateSolution(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

type MockMetricsCollector struct {
	mock.Mock
}

// RecordNodeStatus implements hephaestus.MetricsCollectionService.
func (m *MockMetricsCollector) RecordNodeStatus(nodeID string, status string) {
	panic("unimplemented")
}

func (m *MockMetricsCollector) RecordOperationMetrics(operationName string, duration time.Duration, successful bool) {
	m.Called(operationName, duration, successful)
}

func (m *MockMetricsCollector) RecordErrorMetrics(componentName string, err error) {
	m.Called(componentName, err)
}

func (m *MockMetricsCollector) GetCurrentMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

type MockRepositoryManager struct {
	mock.Mock
}

func (m *MockRepositoryManager) GetFileContents(ctx context.Context, filePath string) ([]byte, error) {
	args := m.Called(ctx, filePath)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRepositoryManager) UpdateFileContents(ctx context.Context, filePath string, contents []byte) error {
	args := m.Called(ctx, filePath, contents)
	return args.Error(0)
}

func (m *MockRepositoryManager) ListRepositoryFiles(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func setupTest(t *testing.T) (*Service, *MockModelClient, *MockMetricsCollector, *MockRepositoryManager, context.Context) {
	// Initialize logger
	err := logger.Initialize(&logger.Config{
		Level:      "debug",
		OutputPath: "",
	})
	require.NoError(t, err)

	// Create mocks
	mockClient := new(MockModelClient)
	mockMetrics := new(MockMetricsCollector)
	mockRepo := new(MockRepositoryManager)

	// Create service
	service := NewService(mockClient, mockMetrics)
	require.NotNil(t, service)

	return service, mockClient, mockMetrics, mockRepo, context.Background()
}

func TestGenerateSolutionProposal(t *testing.T) {
	service, mockClient, mockMetrics, mockRepo, ctx := setupTest(t)

	// Test successful solution generation
	t.Run("successful generation", func(t *testing.T) {
		logEntry := &hephaestus.LogEntryData{
			NodeIdentifier: "test-node",
			LogLevel:       "error",
			LogMessage:     "test error",
			Timestamp:      time.Now(),
			ErrorTrace:     "test stack trace",
		}

		mockClient.On("GenerateSolution", mock.Anything, mock.Anything).Return("test solution", nil)
		mockMetrics.On("RecordOperationMetrics", "generate_solution", mock.Anything, true).Return()

		solution, err := service.GenerateSolutionProposal(ctx, logEntry, mockRepo)
		assert.NoError(t, err)
		assert.NotNil(t, solution)
		assert.Equal(t, logEntry.NodeIdentifier, solution.NodeIdentifier)
		assert.Equal(t, "test solution", solution.ProposedChanges)
		assert.Equal(t, logEntry, solution.AssociatedLog)
	})

	// Test failed solution generation
	t.Run("failed generation", func(t *testing.T) {
		logEntry := &hephaestus.LogEntryData{
			NodeIdentifier: "test-node",
			LogLevel:       "error",
			LogMessage:     "test error",
		}

		mockClient.On("GenerateSolution", mock.Anything, mock.Anything).Return("", assert.AnError)
		mockMetrics.On("RecordOperationMetrics", "generate_solution", mock.Anything, false).Return()
		mockMetrics.On("RecordErrorMetrics", "model_service", assert.AnError).Return()

		solution, err := service.GenerateSolutionProposal(ctx, logEntry, mockRepo)
		assert.Error(t, err)
		assert.Nil(t, solution)
	})
}

func TestValidateSolutionProposal(t *testing.T) {
	service, mockClient, mockMetrics, _, ctx := setupTest(t)

	// Test successful validation
	t.Run("successful validation", func(t *testing.T) {
		solution := &hephaestus.ProposedSolution{
			SolutionID:      "test-solution",
			NodeIdentifier:  "test-node",
			ProposedChanges: "test changes",
			AffectedFiles:   []string{"test.go"},
			GenerationTime:  time.Now(),
		}

		mockClient.On("GenerateSolution", mock.Anything, mock.Anything).Return("valid", nil)
		mockMetrics.On("RecordOperationMetrics", "validate_solution", mock.Anything, true).Return()

		err := service.ValidateSolutionProposal(ctx, solution)
		assert.NoError(t, err)
	})

	// Test failed validation
	t.Run("failed validation", func(t *testing.T) {
		solution := &hephaestus.ProposedSolution{
			SolutionID:      "test-solution",
			NodeIdentifier:  "test-node",
			ProposedChanges: "test changes",
		}

		mockClient.On("GenerateSolution", mock.Anything, mock.Anything).Return("invalid", assert.AnError)
		mockMetrics.On("RecordOperationMetrics", "validate_solution", mock.Anything, false).Return()
		mockMetrics.On("RecordErrorMetrics", "model_service", assert.AnError).Return()

		err := service.ValidateSolutionProposal(ctx, solution)
		assert.Error(t, err)
	})
}

func TestInitialize(t *testing.T) {
	ctx, service, _, _, _ := setupTest(t)

	t.Run("successful initialization", func(t *testing.T) {
		config := &hephaestus.ModelServiceConfiguration{
			ServiceProvider: "test-provider",
			ServiceAPIKey:   "test-key",
			ModelVersion:    "v1",
		}

		err := service.Initialize(ctx, config)
		assert.NoError(t, err)
		assert.Equal(t, config, service.config)
	})

	t.Run("missing configuration", func(t *testing.T) {
		err := service.Initialize(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration is required")
	})

	t.Run("missing API key", func(t *testing.T) {
		config := &hephaestus.ModelServiceConfiguration{
			ServiceProvider: "test-provider",
			ModelVersion:    "v1",
		}

		err := service.Initialize(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key is required")
	})
}

func TestCleanup(t *testing.T) {
	ctx, service, _, _, _ := setupTest(t)

	t.Run("cleanup existing session", func(t *testing.T) {
		nodeID := "test-node"
		session, err := service.getOrCreateSession(ctx, nodeID)
		assert.NoError(t, err)
		assert.NotNil(t, session)

		err = service.Cleanup(ctx, nodeID)
		assert.NoError(t, err)

		_, exists := service.sessions[nodeID]
		assert.False(t, exists)
	})

	t.Run("cleanup non-existent session", func(t *testing.T) {
		err := service.Cleanup(ctx, "non-existent")
		assert.NoError(t, err)
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

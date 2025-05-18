package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ModelServiceClient represents a client for interacting with the model service
type ModelServiceClient interface {
	GenerateSolution(ctx context.Context, prompt string) (string, error)
}

// Service implements the ModelServiceProvider interface
type Service struct {
	config *hephaestus.ModelServiceConfiguration
	client ModelServiceClient
	metricsCollector hephaestus.MetricsCollectionService

	// Active model sessions
	sessions     map[string]*ModelSession
	sessionMutex sync.RWMutex
}

// ModelSession represents an active model interaction session
type ModelSession struct {
	NodeID        string
	LastActive    time.Time
	IsActive      bool
	Configuration *hephaestus.ModelServiceConfiguration
}

// NewService creates a new model service instance
func NewService(client ModelServiceClient, metrics hephaestus.MetricsCollectionService) *Service {
	return &Service{
		client:          client,
		metricsCollector: metrics,
		sessions:         make(map[string]*ModelSession),
	}
}

// Initialize sets up the model service with the provided configuration
func (s *Service) Initialize(ctx context.Context, config *hephaestus.ModelServiceConfiguration) error {
	if config == nil {
		logger.Error(ctx, "model configuration is required")
		return fmt.Errorf("model configuration is required")
	}

	if config.ServiceAPIKey == "" {
		logger.Error(ctx, "model service API key is required")
		return fmt.Errorf("model service API key is required")
	}

	logger.Info(ctx, "initializing model service", 
		zap.String("model_provider", config.ServiceProvider),
		zap.String("model_version", config.ModelVersion),
	)

	s.config = config
	return nil
}

// GenerateSolutionProposal generates a solution proposal for a given log entry
func (s *Service) GenerateSolutionProposal(ctx context.Context, entry *hephaestus.LogEntryData, repo hephaestus.RepositoryManager) (*hephaestus.ProposedSolution, error) {
	start := time.Now()
	logger.Info(ctx, "Generating solution proposal", logger.Field("node_id", entry.NodeIdentifier))

	// Generate prompt from log entry
	prompt := fmt.Sprintf("Log: %s\nLevel: %s\nError: %s\nStack Trace: %s",
		entry.LogMessage,
		entry.LogLevel,
		entry.ErrorTrace,
		entry.ErrorTrace,
	)

	// Call model service
	solution, err := s.client.GenerateSolution(ctx, prompt)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to generate solution", logger.Field("error", err))
		s.metricsCollector.RecordOperationMetrics("generate_solution", duration, false)
		s.metricsCollector.RecordErrorMetrics("model_service", err)
		return nil, fmt.Errorf("failed to generate solution: %v", err)
	}

	// Create solution proposal
	proposal := &hephaestus.ProposedSolution{
		SolutionID:     uuid.New().String(),
		NodeIdentifier: entry.NodeIdentifier,
		AssociatedLog:  entry,
		ProposedChanges: solution,
		GenerationTime: time.Now(),
		ConfidenceScore: 0.8, // TODO: Implement confidence scoring
	}

	s.metricsCollector.RecordOperationMetrics("generate_solution", duration, true)
	logger.Info(ctx, "Generated solution proposal", logger.Field("solution_id", proposal.SolutionID))

	return proposal, nil
}

// ValidateSolutionProposal validates a generated solution proposal
func (s *Service) ValidateSolutionProposal(ctx context.Context, solution *hephaestus.ProposedSolution) error {
	start := time.Now()
	logger.Info(ctx, "Validating solution proposal", logger.Field("solution_id", solution.SolutionID))

	// Generate validation prompt
	prompt := fmt.Sprintf("Validate solution:\n%s", solution.ProposedChanges)

	// Call model service for validation
	result, err := s.client.GenerateSolution(ctx, prompt)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to validate solution", logger.Field("error", err))
		s.metricsCollector.RecordOperationMetrics("validate_solution", duration, false)
		s.metricsCollector.RecordErrorMetrics("model_service", err)
		return fmt.Errorf("failed to validate solution: %v", err)
	}

	// Check validation result
	if result != "valid" {
		err := fmt.Errorf("solution validation failed: %s", result)
		s.metricsCollector.RecordOperationMetrics("validate_solution", duration, false)
		s.metricsCollector.RecordErrorMetrics("model_service", err)
		return err
	}

	s.metricsCollector.RecordOperationMetrics("validate_solution", duration, true)
	logger.Info(ctx, "Solution proposal validated", logger.Field("solution_id", solution.SolutionID))

	return nil
}

// Cleanup removes the model session for a node
func (s *Service) Cleanup(ctx context.Context, nodeID string) error {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	if session, exists := s.sessions[nodeID]; exists {
		logger.Info(ctx, "cleaning up model session",
			zap.String("node_id", nodeID),
			zap.Time("last_active", session.LastActive),
		)
		delete(s.sessions, nodeID)
	} else {
		logger.Warn(ctx, "no model session found for cleanup",
			zap.String("node_id", nodeID),
		)
	}

	return nil
}

// Helper functions

func (s *Service) getOrCreateSession(ctx context.Context, nodeID string) (*ModelSession, error) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	session, exists := s.sessions[nodeID]
	if !exists {
		logger.Info(ctx, "creating new model session",
			zap.String("node_id", nodeID),
		)
		session = &ModelSession{
			NodeID:        nodeID,
			LastActive:    time.Now(),
			IsActive:      true,
			Configuration: s.config,
		}
		s.sessions[nodeID] = session
	} else {
		logger.Debug(ctx, "using existing model session",
			zap.String("node_id", nodeID),
			zap.Time("last_active", session.LastActive),
		)
		session.LastActive = time.Now()
	}

	return session, nil
} 
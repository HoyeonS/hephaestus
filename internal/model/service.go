package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"go.uber.org/zap"
)

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

// ModelServiceClient defines the interface for model service interactions
type ModelServiceClient interface {
	GenerateSolution(ctx context.Context, entry *hephaestus.LogEntryData, config *hephaestus.ModelServiceConfiguration) (*hephaestus.ProposedSolution, error)
}

// NewService creates a new instance of the model service
func NewService(client ModelServiceClient, metricsCollector hephaestus.MetricsCollectionService) *Service {
	return &Service{
		client:           client,
		metricsCollector: metricsCollector,
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

// GenerateSolutionProposal attempts to generate a solution for the given log entry
func (s *Service) GenerateSolutionProposal(ctx context.Context, entry *hephaestus.LogEntryData, repo hephaestus.RepositoryManager) (*hephaestus.ProposedSolution, error) {
	session, err := s.getOrCreateSession(ctx, entry.NodeIdentifier)
	if err != nil {
		logger.Error(ctx, "failed to get or create model session",
			zap.String("node_id", entry.NodeIdentifier),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get or create model session: %v", err)
	}

	startTime := time.Now()
	logger.Info(ctx, "generating solution",
		zap.String("node_id", entry.NodeIdentifier),
		zap.String("log_level", entry.LogLevel),
	)

	solution, err := s.client.GenerateSolution(ctx, entry, s.config)
	if err != nil {
		logger.Error(ctx, "failed to generate solution",
			zap.String("node_id", entry.NodeIdentifier),
			zap.String("log_level", entry.LogLevel),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate solution: %v", err)
	}

	// Record latency metric
	latency := time.Since(startTime)
	if err := s.metricsCollector.RecordOperationMetrics("generate_solution", latency, true); err != nil {
		logger.Warn(ctx, "failed to record model latency metric",
			zap.String("node_id", entry.NodeIdentifier),
			zap.Duration("latency", latency),
			zap.Error(err),
		)
	}

	logger.Info(ctx, "solution generated successfully",
		zap.String("node_id", entry.NodeIdentifier),
		zap.String("log_level", entry.LogLevel),
		zap.Duration("duration", latency),
	)

	return solution, nil
}

// ValidateSolutionProposal validates a generated solution
func (s *Service) ValidateSolutionProposal(ctx context.Context, solution *hephaestus.ProposedSolution) error {
	if solution == nil {
		return fmt.Errorf("solution cannot be nil")
	}

	logger.Info(ctx, "validating solution proposal",
		zap.String("node_id", solution.NodeIdentifier),
		zap.Float64("confidence", solution.ConfidenceScore),
	)

	// Add validation logic here
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
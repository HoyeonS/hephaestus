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
	config *hephaestus.ModelSettings
	client ModelClient
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
	Configuration *hephaestus.ModelSettings
}

// NewService creates a new instance of the model service
func NewService(client ModelClient, metricsCollector hephaestus.MetricsCollectionService) *Service {
	return &Service{
		client:           client,
		metricsCollector: metricsCollector,
		sessions:         make(map[string]*ModelSession),
	}
}

// Initialize sets up the model service with the provided configuration
func (s *Service) Initialize(ctx context.Context, config *hephaestus.ModelSettings) error {
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
		zap.String("model_name", config.ModelName),
	)

	s.config = config
	return nil
}

// GenerateSolution attempts to generate a solution for the given problem
func (s *Service) GenerateSolution(ctx context.Context, nodeID string, problem *hephaestus.Problem) (*hephaestus.Solution, error) {
	session, err := s.getOrCreateSession(ctx, nodeID)
	if err != nil {
		logger.Error(ctx, "failed to get or create model session",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get or create model session: %v", err)
	}

	startTime := time.Now()
	logger.Info(ctx, "generating solution",
		zap.String("node_id", nodeID),
		zap.String("problem_type", problem.Type),
	)

	solution, err := s.client.GenerateSolution(ctx, problem, s.config)
	if err != nil {
		logger.Error(ctx, "failed to generate solution",
			zap.String("node_id", nodeID),
			zap.String("problem_type", problem.Type),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate solution: %v", err)
	}

	// Record latency metric
	latency := time.Since(startTime)
	if err := s.metricsCollector.RecordModelLatency(ctx, nodeID, latency); err != nil {
		logger.Warn(ctx, "failed to record model latency metric",
			zap.String("node_id", nodeID),
			zap.Duration("latency", latency),
			zap.Error(err),
		)
	}

	logger.Info(ctx, "solution generated successfully",
		zap.String("node_id", nodeID),
		zap.String("problem_type", problem.Type),
		zap.Duration("duration", latency),
	)

	return solution, nil
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
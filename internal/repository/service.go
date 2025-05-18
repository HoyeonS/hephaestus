package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// Service implements the RepositoryManager interface
type Service struct {
	remoteService     hephaestus.RemoteRepositoryService
	metricsCollector  hephaestus.MetricsCollectionService
}

// NewService creates a new repository service instance
func NewService(remote hephaestus.RemoteRepositoryService, metrics hephaestus.MetricsCollectionService) *Service {
	return &Service{
		remoteService:     remote,
		metricsCollector:  metrics,
	}
}

// Initialize sets up the repository service
func (s *Service) Initialize(ctx context.Context) error {
	logger.Info(ctx, "Initializing repository service")

	if err := s.remoteService.Initialize(ctx); err != nil {
		logger.Error(ctx, "Failed to initialize remote service", logger.Field("error", err))
		return fmt.Errorf("failed to initialize remote service: %v", err)
	}

	return nil
}

// GetFileContents retrieves the contents of a file from the repository
func (s *Service) GetFileContents(ctx context.Context, path string) ([]byte, error) {
	start := time.Now()
	logger.Info(ctx, "Retrieving file contents", logger.Field("path", path))

	content, err := s.remoteService.GetFileContents(ctx, path)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to retrieve file contents",
			logger.Field("path", path),
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("get_file_contents", duration, false)
		return nil, fmt.Errorf("failed to retrieve file contents: %v", err)
	}

	s.metricsCollector.RecordOperationMetrics("get_file_contents", duration, true)
	return content, nil
}

// UpdateFileContents updates the contents of a file in the repository
func (s *Service) UpdateFileContents(ctx context.Context, path string, content []byte, message string) error {
	start := time.Now()
	logger.Info(ctx, "Updating file contents",
		logger.Field("path", path),
		logger.Field("message", message),
	)

	err := s.remoteService.UpdateFileContents(ctx, path, content, message)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to update file contents",
			logger.Field("path", path),
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("update_file_contents", duration, false)
		return fmt.Errorf("failed to update file contents: %v", err)
	}

	s.metricsCollector.RecordOperationMetrics("update_file_contents", duration, true)
	return nil
}

// CreatePullRequest creates a new pull request in the repository
func (s *Service) CreatePullRequest(ctx context.Context, title, body, head, base string) (string, error) {
	start := time.Now()
	logger.Info(ctx, "Creating pull request",
		logger.Field("title", title),
		logger.Field("head", head),
		logger.Field("base", base),
	)

	url, err := s.remoteService.CreatePullRequest(ctx, title, body, head, base)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to create pull request",
			logger.Field("title", title),
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("create_pull_request", duration, false)
		return "", fmt.Errorf("failed to create pull request: %v", err)
	}

	s.metricsCollector.RecordOperationMetrics("create_pull_request", duration, true)
	return url, nil
} 
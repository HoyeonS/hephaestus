package remote

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/google/go-github/v45/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// Service implements the RemoteRepositoryService interface
type Service struct {
	client           *github.Client
	metricsCollector hephaestus.MetricsCollectionService
	owner            string
	repo             string

	// Active repository connections
	connections     map[string]*RepositoryConnection
	connectionMutex sync.RWMutex
}

type RepositoryConnection struct {
	NodeID        string
	Owner         string
	Repository    string
	LastActive    time.Time
	IsActive      bool
}

// NewService creates a new remote repository service instance
func NewService(token, owner, repo string, metrics hephaestus.MetricsCollectionService) *Service {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Service{
		client:           github.NewClient(tc),
		metricsCollector: metrics,
		owner:            owner,
		repo:             repo,
		connections:      make(map[string]*RepositoryConnection),
	}
}

// Initialize sets up the remote repository service
func (s *Service) Initialize(ctx context.Context) error {
	logger.Info(ctx, "Initializing remote repository service",
		logger.Field("owner", s.owner),
		logger.Field("repo", s.repo),
	)

	// Verify repository access
	_, _, err := s.client.Repositories.Get(ctx, s.owner, s.repo)
	if err != nil {
		logger.Error(ctx, "Failed to access repository",
			logger.Field("error", err),
		)
		return fmt.Errorf("failed to access repository: %v", err)
	}

	return nil
}

// GetFileContents retrieves the contents of a file from the repository
func (s *Service) GetFileContents(ctx context.Context, path string) ([]byte, error) {
	start := time.Now()
	logger.Info(ctx, "Retrieving file contents",
		logger.Field("path", path),
	)

	fileContent, _, _, err := s.client.Repositories.GetContents(ctx, s.owner, s.repo, path, nil)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to retrieve file contents",
			logger.Field("path", path),
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("get_file_contents", duration, false)
		return nil, fmt.Errorf("failed to retrieve file contents: %v", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		logger.Error(ctx, "Failed to decode file contents",
			logger.Field("path", path),
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("get_file_contents", duration, false)
		return nil, fmt.Errorf("failed to decode file contents: %v", err)
	}

	s.metricsCollector.RecordOperationMetrics("get_file_contents", duration, true)
	return []byte(content), nil
}

// UpdateFileContents updates the contents of a file in the repository
func (s *Service) UpdateFileContents(ctx context.Context, path string, content []byte, message string) error {
	start := time.Now()
	logger.Info(ctx, "Updating file contents",
		logger.Field("path", path),
	)

	// Get the current file to obtain its SHA
	fileContent, _, _, err := s.client.Repositories.GetContents(ctx, s.owner, s.repo, path, nil)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to get current file contents",
			logger.Field("path", path),
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("update_file_contents", duration, false)
		return fmt.Errorf("failed to get current file contents: %v", err)
	}

	// Create the update request
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: content,
		SHA:     fileContent.SHA,
	}

	_, _, err = s.client.Repositories.UpdateFile(ctx, s.owner, s.repo, path, opts)
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

	pr := &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	}

	pullRequest, _, err := s.client.PullRequests.Create(ctx, s.owner, s.repo, pr)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Failed to create pull request",
			logger.Field("error", err),
		)
		s.metricsCollector.RecordOperationMetrics("create_pull_request", duration, false)
		return "", fmt.Errorf("failed to create pull request: %v", err)
	}

	s.metricsCollector.RecordOperationMetrics("create_pull_request", duration, true)
	return pullRequest.GetHTMLURL(), nil
}

// CreateIssue creates a new issue in the remote repository
func (s *Service) CreateIssue(ctx context.Context, nodeID string, issue *hephaestus.Issue) error {
	conn, err := s.getOrCreateConnection(ctx, nodeID)
	if err != nil {
		logger.Error(ctx, "failed to get or create repository connection",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get or create repository connection: %v", err)
	}

	logger.Info(ctx, "creating new issue",
		zap.String("node_id", nodeID),
		zap.String("title", issue.Title),
	)

	// Create GitHub issue
	githubIssue := &github.IssueRequest{
		Title:     &issue.Title,
		Body:      &issue.Description,
		Labels:    &issue.Labels,
		Assignees: &issue.Assignees,
	}

	_, _, err = s.client.Issues.Create(ctx, conn.Owner, conn.Repository, githubIssue)
	if err != nil {
		logger.Error(ctx, "failed to create issue",
			zap.String("node_id", nodeID),
			zap.String("title", issue.Title),
			zap.Error(err),
		)
		if err := s.metricsCollector.RecordRepositoryError(ctx, nodeID, "create_issue"); err != nil {
			logger.Warn(ctx, "failed to record repository error metric",
				zap.String("node_id", nodeID),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to create issue: %v", err)
	}

	logger.Info(ctx, "issue created successfully",
		zap.String("node_id", nodeID),
		zap.String("title", issue.Title),
	)

	return nil
}

// Cleanup removes the repository connection for a node
func (s *Service) Cleanup(ctx context.Context, nodeID string) error {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()

	if conn, exists := s.connections[nodeID]; exists {
		logger.Info(ctx, "cleaning up repository connection",
			zap.String("node_id", nodeID),
			zap.Time("last_active", conn.LastActive),
		)
		delete(s.connections, nodeID)
	} else {
		logger.Warn(ctx, "no repository connection found for cleanup",
			zap.String("node_id", nodeID),
		)
	}

	return nil
}

// Helper functions

func (s *Service) getOrCreateConnection(ctx context.Context, nodeID string) (*RepositoryConnection, error) {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()

	conn, exists := s.connections[nodeID]
	if !exists {
		logger.Info(ctx, "creating new repository connection",
			zap.String("node_id", nodeID),
			zap.String("owner", s.owner),
			zap.String("repository", s.repo),
		)
		conn = &RepositoryConnection{
			NodeID:     nodeID,
			Owner:      s.owner,
			Repository: s.repo,
			LastActive: time.Now(),
			IsActive:   true,
		}
		s.connections[nodeID] = conn
	} else {
		logger.Debug(ctx, "using existing repository connection",
			zap.String("node_id", nodeID),
			zap.Time("last_active", conn.LastActive),
		)
		conn.LastActive = time.Now()
	}

	return conn, nil
} 
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
	client          *github.Client
	config          *hephaestus.RemoteSettings
	metricsCollector hephaestus.MetricsCollectionService

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

// NewService creates a new instance of the remote repository service
func NewService(metricsCollector hephaestus.MetricsCollectionService) *Service {
	return &Service{
		metricsCollector: metricsCollector,
		connections:     make(map[string]*RepositoryConnection),
	}
}

// Initialize sets up the remote repository service with the provided configuration
func (s *Service) Initialize(ctx context.Context, config *hephaestus.RemoteSettings) error {
	if config == nil {
		logger.Error(ctx, "remote repository configuration is required")
		return fmt.Errorf("remote repository configuration is required")
	}

	if config.AuthToken == "" {
		logger.Error(ctx, "authentication token is required")
		return fmt.Errorf("authentication token is required")
	}

	logger.Info(ctx, "initializing remote repository service",
		zap.String("owner", config.RepositoryOwner),
		zap.String("repository", config.RepositoryName),
	)

	// Create GitHub client with authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.AuthToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	s.client = github.NewClient(tc)
	s.config = config

	// Validate repository access
	if err := s.validateRepositoryAccess(ctx); err != nil {
		logger.Error(ctx, "failed to validate repository access",
			zap.String("owner", config.RepositoryOwner),
			zap.String("repository", config.RepositoryName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to validate repository access: %v", err)
	}

	return nil
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

// CreatePullRequest creates a new pull request in the remote repository
func (s *Service) CreatePullRequest(ctx context.Context, nodeID string, pr *hephaestus.PullRequest) error {
	conn, err := s.getOrCreateConnection(ctx, nodeID)
	if err != nil {
		logger.Error(ctx, "failed to get or create repository connection",
			zap.String("node_id", nodeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get or create repository connection: %v", err)
	}

	logger.Info(ctx, "creating new pull request",
		zap.String("node_id", nodeID),
		zap.String("title", pr.Title),
		zap.String("base", pr.BaseBranch),
		zap.String("head", pr.HeadBranch),
	)

	// Create GitHub pull request
	githubPR := &github.NewPullRequest{
		Title:               &pr.Title,
		Body:               &pr.Description,
		Head:               &pr.HeadBranch,
		Base:               &pr.BaseBranch,
		MaintainerCanModify: github.Bool(true),
	}

	_, _, err = s.client.PullRequests.Create(ctx, conn.Owner, conn.Repository, githubPR)
	if err != nil {
		logger.Error(ctx, "failed to create pull request",
			zap.String("node_id", nodeID),
			zap.String("title", pr.Title),
			zap.Error(err),
		)
		if err := s.metricsCollector.RecordRepositoryError(ctx, nodeID, "create_pr"); err != nil {
			logger.Warn(ctx, "failed to record repository error metric",
				zap.String("node_id", nodeID),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to create pull request: %v", err)
	}

	logger.Info(ctx, "pull request created successfully",
		zap.String("node_id", nodeID),
		zap.String("title", pr.Title),
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

func (s *Service) validateRepositoryAccess(ctx context.Context) error {
	logger.Debug(ctx, "validating repository access",
		zap.String("owner", s.config.RepositoryOwner),
		zap.String("repository", s.config.RepositoryName),
	)

	_, _, err := s.client.Repositories.Get(ctx, s.config.RepositoryOwner, s.config.RepositoryName)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getOrCreateConnection(ctx context.Context, nodeID string) (*RepositoryConnection, error) {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()

	conn, exists := s.connections[nodeID]
	if !exists {
		logger.Info(ctx, "creating new repository connection",
			zap.String("node_id", nodeID),
			zap.String("owner", s.config.RepositoryOwner),
			zap.String("repository", s.config.RepositoryName),
		)
		conn = &RepositoryConnection{
			NodeID:     nodeID,
			Owner:      s.config.RepositoryOwner,
			Repository: s.config.RepositoryName,
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
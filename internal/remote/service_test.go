package remote

import (
	"context"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	args := m.Called(ctx, owner, repo)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*github.Repository), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) Create(ctx context.Context, owner, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	args := m.Called(ctx, owner, repo, issue)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*github.Issue), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) CreatePR(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	args := m.Called(ctx, owner, repo, pr)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*github.PullRequest), args.Get(1).(*github.Response), args.Error(2)
}

type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordRepositoryError(ctx context.Context, nodeID string, operation string) error {
	args := m.Called(ctx, nodeID, operation)
	return args.Error(0)
}

func setupTest(t *testing.T) (context.Context, *Service, *MockGitHubClient, *MockMetricsCollector) {
	// Initialize logger for tests
	err := logger.Initialize(&logger.Config{
		Level:  "debug",
		Format: "console",
	})
	require.NoError(t, err)

	ctx := context.Background()
	mockClient := new(MockGitHubClient)
	mockMetrics := new(MockMetricsCollector)
	service := NewService(mockMetrics)
	service.client = &github.Client{
		Repositories: &github.RepositoriesService{},
		Issues:      &github.IssuesService{},
	}

	return ctx, service, mockClient, mockMetrics
}

func TestInitialize(t *testing.T) {
	ctx, service, mockClient, _ := setupTest(t)

	t.Run("successful initialization", func(t *testing.T) {
		config := &hephaestus.RemoteSettings{
			AuthToken:       "test-token",
			RepositoryOwner: "test-owner",
			RepositoryName:  "test-repo",
		}

		mockClient.On("Get", ctx, config.RepositoryOwner, config.RepositoryName).Return(&github.Repository{}, &github.Response{}, nil)

		err := service.Initialize(ctx, config)
		require.NoError(t, err)
		assert.Equal(t, config, service.config)

		mockClient.AssertExpectations(t)
	})

	t.Run("nil configuration", func(t *testing.T) {
		err := service.Initialize(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "remote repository configuration is required")
	})

	t.Run("missing auth token", func(t *testing.T) {
		config := &hephaestus.RemoteSettings{
			RepositoryOwner: "test-owner",
			RepositoryName:  "test-repo",
		}

		err := service.Initialize(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication token is required")
	})

	t.Run("invalid repository access", func(t *testing.T) {
		config := &hephaestus.RemoteSettings{
			AuthToken:       "test-token",
			RepositoryOwner: "test-owner",
			RepositoryName:  "test-repo",
		}

		mockClient.On("Get", ctx, config.RepositoryOwner, config.RepositoryName).Return(nil, nil, assert.AnError)

		err := service.Initialize(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate repository access")

		mockClient.AssertExpectations(t)
	})
}

func TestCreateIssue(t *testing.T) {
	ctx, service, mockClient, mockMetrics := setupTest(t)

	config := &hephaestus.RemoteSettings{
		AuthToken:       "test-token",
		RepositoryOwner: "test-owner",
		RepositoryName:  "test-repo",
	}
	err := service.Initialize(ctx, config)
	require.NoError(t, err)

	t.Run("successful issue creation", func(t *testing.T) {
		nodeID := "test-node"
		issue := &hephaestus.Issue{
			Title:       "Test Issue",
			Description: "Test Description",
			Labels:      []string{"bug"},
			Assignees:   []string{"user1"},
		}

		githubIssue := &github.IssueRequest{
			Title:     &issue.Title,
			Body:      &issue.Description,
			Labels:    &issue.Labels,
			Assignees: &issue.Assignees,
		}

		mockClient.On("Create", ctx, config.RepositoryOwner, config.RepositoryName, githubIssue).Return(&github.Issue{}, &github.Response{}, nil)

		err := service.CreateIssue(ctx, nodeID, issue)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)

		// Verify connection was created
		conn, exists := service.connections[nodeID]
		assert.True(t, exists)
		assert.Equal(t, nodeID, conn.NodeID)
		assert.Equal(t, config.RepositoryOwner, conn.Owner)
		assert.Equal(t, config.RepositoryName, conn.Repository)
		assert.True(t, conn.IsActive)
	})

	t.Run("failed issue creation", func(t *testing.T) {
		nodeID := "test-node"
		issue := &hephaestus.Issue{
			Title:       "Test Issue",
			Description: "Test Description",
			Labels:      []string{"bug"},
			Assignees:   []string{"user1"},
		}

		githubIssue := &github.IssueRequest{
			Title:     &issue.Title,
			Body:      &issue.Description,
			Labels:    &issue.Labels,
			Assignees: &issue.Assignees,
		}

		mockClient.On("Create", ctx, config.RepositoryOwner, config.RepositoryName, githubIssue).Return(nil, nil, assert.AnError)
		mockMetrics.On("RecordRepositoryError", ctx, nodeID, "create_issue").Return(nil)

		err := service.CreateIssue(ctx, nodeID, issue)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create issue")

		mockClient.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("failed metrics recording", func(t *testing.T) {
		nodeID := "test-node"
		issue := &hephaestus.Issue{
			Title:       "Test Issue",
			Description: "Test Description",
			Labels:      []string{"bug"},
			Assignees:   []string{"user1"},
		}

		githubIssue := &github.IssueRequest{
			Title:     &issue.Title,
			Body:      &issue.Description,
			Labels:    &issue.Labels,
			Assignees: &issue.Assignees,
		}

		mockClient.On("Create", ctx, config.RepositoryOwner, config.RepositoryName, githubIssue).Return(nil, nil, assert.AnError)
		mockMetrics.On("RecordRepositoryError", ctx, nodeID, "create_issue").Return(assert.AnError)

		err := service.CreateIssue(ctx, nodeID, issue)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create issue")

		mockClient.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

func TestCreatePullRequest(t *testing.T) {
	ctx, service, mockClient, mockMetrics := setupTest(t)

	config := &hephaestus.RemoteSettings{
		AuthToken:       "test-token",
		RepositoryOwner: "test-owner",
		RepositoryName:  "test-repo",
	}
	err := service.Initialize(ctx, config)
	require.NoError(t, err)

	t.Run("successful pull request creation", func(t *testing.T) {
		nodeID := "test-node"
		pr := &hephaestus.PullRequest{
			Title:       "Test PR",
			Description: "Test Description",
			BaseBranch:  "main",
			HeadBranch:  "feature",
		}

		githubPR := &github.NewPullRequest{
			Title:               &pr.Title,
			Body:               &pr.Description,
			Head:               &pr.HeadBranch,
			Base:               &pr.BaseBranch,
			MaintainerCanModify: github.Bool(true),
		}

		mockClient.On("CreatePR", ctx, config.RepositoryOwner, config.RepositoryName, githubPR).Return(&github.PullRequest{}, &github.Response{}, nil)

		err := service.CreatePullRequest(ctx, nodeID, pr)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("failed pull request creation", func(t *testing.T) {
		nodeID := "test-node"
		pr := &hephaestus.PullRequest{
			Title:       "Test PR",
			Description: "Test Description",
			BaseBranch:  "main",
			HeadBranch:  "feature",
		}

		githubPR := &github.NewPullRequest{
			Title:               &pr.Title,
			Body:               &pr.Description,
			Head:               &pr.HeadBranch,
			Base:               &pr.BaseBranch,
			MaintainerCanModify: github.Bool(true),
		}

		mockClient.On("CreatePR", ctx, config.RepositoryOwner, config.RepositoryName, githubPR).Return(nil, nil, assert.AnError)
		mockMetrics.On("RecordRepositoryError", ctx, nodeID, "create_pr").Return(nil)

		err := service.CreatePullRequest(ctx, nodeID, pr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create pull request")

		mockClient.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

func TestCleanup(t *testing.T) {
	ctx, service, _, _ := setupTest(t)

	config := &hephaestus.RemoteSettings{
		AuthToken:       "test-token",
		RepositoryOwner: "test-owner",
		RepositoryName:  "test-repo",
	}
	err := service.Initialize(ctx, config)
	require.NoError(t, err)

	t.Run("cleanup existing connection", func(t *testing.T) {
		nodeID := "test-node"
		conn := &RepositoryConnection{
			NodeID:     nodeID,
			Owner:      config.RepositoryOwner,
			Repository: config.RepositoryName,
			LastActive: time.Now(),
			IsActive:   true,
		}
		service.connections[nodeID] = conn

		err := service.Cleanup(ctx, nodeID)
		require.NoError(t, err)
		assert.NotContains(t, service.connections, nodeID)
	})

	t.Run("cleanup non-existent connection", func(t *testing.T) {
		err := service.Cleanup(ctx, "non-existent")
		require.NoError(t, err)
	})
} 
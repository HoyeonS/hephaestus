package repository

import (
	"context"
	"testing"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock GitHub client
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	args := m.Called(ctx, owner, repo)
	return args.Get(0).(*github.Repository), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) GetBranch(ctx context.Context, owner, repo, branch string) (*github.Branch, *github.Response, error) {
	args := m.Called(ctx, owner, repo, branch)
	return args.Get(0).(*github.Branch), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) CreateRef(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	args := m.Called(ctx, owner, repo, ref)
	return args.Get(0).(*github.Reference), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	args := m.Called(ctx, owner, repo, path, opts)
	return args.Get(0).(*github.RepositoryContent), args.Get(1).([]*github.RepositoryContent), args.Get(2).(*github.Response), args.Error(3)
}

func (m *MockGitHubClient) CreateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
	args := m.Called(ctx, owner, repo, path, opts)
	return args.Get(0).(*github.RepositoryContentResponse), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitHubClient) CreatePullRequest(ctx context.Context, owner, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	args := m.Called(ctx, owner, repo, pull)
	return args.Get(0).(*github.PullRequest), args.Get(1).(*github.Response), args.Error(2)
}

// Test setup helper
func setupTestRemoteService() (*RemoteService, *MockGitHubClient) {
	mockGitHubClient := new(MockGitHubClient)
	service := NewRemoteService()
	service.remoteRepositoryClient = Github.Client(mockGitHubClient)
	return service, mockGitHubClient
}

func TestInitialize(t *testing.T) {
	service, mockGitHubClient := setupTestRemoteService()
	ctx := context.Background()

	validConfig := hephaestus.RemoteRepositoryConfiguration{
		AuthToken:       "test-token",
		RepositoryOwner: "test-owner",
		RepositoryName:  "test-repo",
		TargetBranch:    "main",
	}

	t.Run("Successful initialization", func(t *testing.T) {
		mockGitHubClient.On("GetRepository", ctx, validConfig.RepositoryOwner, validConfig.RepositoryName).
			Return(&github.Repository{Archived: github.Bool(false)}, &github.Response{}, nil)

		err := service.Initialize(ctx, validConfig)

		assert.NoError(t, err)
		assert.Equal(t, &validConfig, service.config)
	})

	t.Run("Missing auth token", func(t *testing.T) {
		invalidConfig := validConfig
		invalidConfig.AuthToken = ""

		err := service.Initialize(ctx, invalidConfig)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth token is required")
	})

	t.Run("Missing repository owner", func(t *testing.T) {
		invalidConfig := validConfig
		invalidConfig.RepositoryOwner = ""

		err := service.Initialize(ctx, invalidConfig)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository owner and name are required")
	})

	t.Run("Archived repository", func(t *testing.T) {
		mockGitHubClient.On("GetRepository", ctx, validConfig.RepositoryOwner, validConfig.RepositoryName).
			Return(&github.Repository{Archived: github.Bool(true)}, &github.Response{}, nil)

		err := service.Initialize(ctx, validConfig)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository is archived")
	})
}

func TestCreateRepository(t *testing.T) {
	service, mockGitHubClient := setupTestRemoteService()
	ctx := context.Background()
	nodeID := "test-node"

	service.config = &hephaestus.RemoteRepositoryConfiguration{
		RepositoryOwner: "test-owner",
		RepositoryName:  "test-repo",
		TargetBranch:    "main",
	}

	t.Run("Successful repository creation", func(t *testing.T) {
		mockGitHubClient.On("GetBranch", ctx, service.config.RepositoryOwner, service.config.RepositoryName, service.config.TargetBranch, true).
			Return(&github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("test-sha"),
				},
			}, &github.Response{}, nil)

		err := service.CreateRepository(ctx, nodeID)

		assert.NoError(t, err)
		repo, exists := service.repositories[nodeID]
		assert.True(t, exists)
		assert.Equal(t, service.config.RepositoryOwner, repo.Owner)
		assert.Equal(t, service.config.RepositoryName, repo.Name)
		assert.Equal(t, service.config.TargetBranch, repo.Branch)
		assert.Equal(t, "test-sha", repo.LastCommit)
	})

	t.Run("Repository already exists", func(t *testing.T) {
		err := service.CreateRepository(ctx, nodeID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository already exists for node")
	})
}

func TestGetRepository(t *testing.T) {
	service, _ := setupTestRemoteService()
	ctx := context.Background()
	nodeID := "test-node"

	t.Run("Get existing repository", func(t *testing.T) {
		service.repositories[nodeID] = &Repository{
			Owner: "test-owner",
			Name:  "test-repo",
		}

		repo, err := service.GetRepository(ctx, nodeID)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, "test-owner", repo.Owner)
		assert.Equal(t, "test-repo", repo.Name)
	})

	t.Run("Get non-existent repository", func(t *testing.T) {
		repo, err := service.GetRepository(ctx, "non-existent")

		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.Contains(t, err.Error(), "repository not found for node")
	})
}

func TestUpdateRepository(t *testing.T) {
	service, mockGitHubClient := setupTestRemoteService()
	ctx := context.Background()
	nodeID := "test-node"

	service.repositories[nodeID] = &Repository{
		Owner:      "test-owner",
		Name:       "test-repo",
		Branch:     "main",
		LastCommit: "old-sha",
	}

	t.Run("Successful update", func(t *testing.T) {
		mockGitHubClient.On("GetRepository", ctx, "test-owner", "test-repo").
			Return(&github.Repository{Archived: github.Bool(false)}, &github.Response{}, nil)
		mockGitHubClient.On("GetBranch", ctx, "test-owner", "test-repo", "main", true).
			Return(&github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("new-sha"),
				},
			}, &github.Response{}, nil)

		err := service.UpdateRepository(ctx, nodeID)

		assert.NoError(t, err)
		assert.Equal(t, "new-sha", service.repositories[nodeID].LastCommit)
	})

	t.Run("Repository archived", func(t *testing.T) {
		mockGitHubClient.On("GetRepository", ctx, "test-owner", "test-repo").
			Return(&github.Repository{Archived: github.Bool(true)}, &github.Response{}, nil)

		err := service.UpdateRepository(ctx, nodeID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository has been archived")
		assert.True(t, service.repositories[nodeID].IsArchived)
	})

	t.Run("Non-existent repository", func(t *testing.T) {
		err := service.UpdateRepository(ctx, "non-existent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not found for node")
	})
}

func TestCreatePullRequest(t *testing.T) {
	service, mockGitHubClient := setupTestRemoteService()
	ctx := context.Background()
	nodeID := "test-node"

	service.repositories[nodeID] = &Repository{
		Owner:      "test-owner",
		Name:       "test-repo",
		Branch:     "main",
		LastCommit: "test-sha",
		IsArchived: false,
	}

	changes := []hephaestus.CodeChange{
		{
			FilePath:    "test.go",
			UpdatedCode: "package main",
		},
	}

	t.Run("Successful pull request creation", func(t *testing.T) {
		mockGitHubClient.On("CreateRef", ctx, "test-owner", "test-repo", mock.Anything).
			Return(&github.Reference{}, &github.Response{}, nil)
		mockGitHubClient.On("GetContents", ctx, "test-owner", "test-repo", "test.go", mock.Anything).
			Return(&github.RepositoryContent{SHA: github.String("content-sha")}, nil, &github.Response{}, nil)
		mockGitHubClient.On("CreateFile", ctx, "test-owner", "test-repo", "test.go", mock.Anything).
			Return(&github.RepositoryContentResponse{}, &github.Response{}, nil)
		mockGitHubClient.On("CreatePullRequest", ctx, "test-owner", "test-repo", mock.Anything).
			Return(&github.PullRequest{}, &github.Response{}, nil)

		err := service.CreatePullRequest(ctx, nodeID, "Test PR", "Test description", changes)

		assert.NoError(t, err)
	})

	t.Run("Archived repository", func(t *testing.T) {
		service.repositories[nodeID].IsArchived = true

		err := service.CreatePullRequest(ctx, nodeID, "Test PR", "Test description", changes)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot create pull request for archived repository")
	})

	t.Run("Non-existent repository", func(t *testing.T) {
		err := service.CreatePullRequest(ctx, "non-existent", "Test PR", "Test description", changes)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not found for node")
	})
}

func TestCleanup(t *testing.T) {
	service, _ := setupTestRemoteService()
	ctx := context.Background()
	nodeID := "test-node"

	t.Run("Successful cleanup", func(t *testing.T) {
		service.repositories[nodeID] = &Repository{
			Owner: "test-owner",
			Name:  "test-repo",
		}

		err := service.Cleanup(ctx, nodeID)

		assert.NoError(t, err)
		_, exists := service.repositories[nodeID]
		assert.False(t, exists)
	})

	t.Run("Non-existent repository", func(t *testing.T) {
		err := service.Cleanup(ctx, "non-existent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not found for node")
	})
}

package repository

import (
	"context"
	"fmt"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// RemoteService implements the RemoteRepositoryService interface
type RemoteService struct {
	// Dependencies
	remoteRepositoryClient *github.Client

	// Configuration
	config *hephaestus.RemoteRepositoryConfiguration

	// Active repositories
	repository Repository
}

// Repository represents a remote repository instance
type Repository struct {
	Owner              string
	Name               string
	Branch             string
	IsArchived         bool
	RemoteFileNodeList []RemoteFileNode
}

type RemoteFileNode struct {
	FilePath string
	Content  string
}

// NewRemoteService creates a new instance of the remote repository service
func NewRemoteService() *RemoteService {
	return &RemoteService{}
}

// Initialize sets up the remote repository service with the provided configuration
func (s *RemoteService) Initialize(ctx context.Context, config hephaestus.RemoteRepositoryConfiguration) error {
	if config.ProviderToken == "" {
		return fmt.Errorf("auth token is required")
	}

	if config.RemoteRepositoryOwner == "" || config.RemoteRepositoryName == "" {
		return fmt.Errorf("repository owner and name are required")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.ProviderToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	s.remoteRepositoryClient = github.NewClient(tc)

	// Verify repository access
	repo, _, err := s.remoteRepositoryClient.Repositories.Get(ctx, config.RemoteRepositoryOwner, config.RemoteRepositoryName)
	if err != nil {
		return fmt.Errorf("failed to access repository: %v", err)
	}

	if repo.GetArchived() {
		return fmt.Errorf("repository is archived")
	}

	s.config = &config
	return nil
}

// CreateRepository creates a new repository instance for a node
func (s *RemoteService) CreateRepository(ctx context.Context, nodeID string) error {

	// Create repository instance
	repo := &Repository{
		Owner:              s.config.RemoteRepositoryOwner,
		Name:               s.config.RemoteRepositoryName,
		Branch:             s.config.RemoteRepositoryBranch,
		RemoteFileNodeList: make([]RemoteFileNode, 0),
		IsArchived:         false,
	}

	s.repository = *repo

	return nil
}

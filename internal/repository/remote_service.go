package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// RemoteService implements the RemoteRepositoryService interface
type RemoteService struct {
	// Dependencies
	githubClient *github.Client
	
	// Configuration
	config *hephaestus.RemoteRepositoryConfiguration
	
	// Active repositories
	repositories     map[string]*Repository
	repositoryMutex sync.RWMutex
}

// Repository represents a remote repository instance
type Repository struct {
	Owner       string
	Name        string
	Branch      string
	LastCommit  string
	WorkingDir  string
	IsArchived  bool
}

// NewRemoteService creates a new instance of the remote repository service
func NewRemoteService() *RemoteService {
	return &RemoteService{
		repositories: make(map[string]*Repository),
	}
}

// Initialize sets up the remote repository service with the provided configuration
func (s *RemoteService) Initialize(ctx context.Context, config hephaestus.RemoteRepositoryConfiguration) error {
	if config.AuthToken == "" {
		return fmt.Errorf("auth token is required")
	}

	if config.RepositoryOwner == "" || config.RepositoryName == "" {
		return fmt.Errorf("repository owner and name are required")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.AuthToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	s.githubClient = github.NewClient(tc)

	// Verify repository access
	repo, _, err := s.githubClient.Repositories.Get(ctx, config.RepositoryOwner, config.RepositoryName)
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
	s.repositoryMutex.Lock()
	defer s.repositoryMutex.Unlock()

	if _, exists := s.repositories[nodeID]; exists {
		return fmt.Errorf("repository already exists for node: %s", nodeID)
	}

	// Create repository instance
	repo := &Repository{
		Owner:      s.config.RepositoryOwner,
		Name:       s.config.RepositoryName,
		Branch:     s.config.TargetBranch,
		IsArchived: false,
	}

	// Get latest commit
	branch, _, err := s.githubClient.Repositories.GetBranch(ctx, repo.Owner, repo.Name, repo.Branch, true)
	if err != nil {
		return fmt.Errorf("failed to get branch information: %v", err)
	}

	repo.LastCommit = branch.GetCommit().GetSHA()
	s.repositories[nodeID] = repo

	return nil
}

// GetRepository retrieves a repository instance for a node
func (s *RemoteService) GetRepository(ctx context.Context, nodeID string) (*Repository, error) {
	s.repositoryMutex.RLock()
	defer s.repositoryMutex.RUnlock()

	repo, exists := s.repositories[nodeID]
	if !exists {
		return nil, fmt.Errorf("repository not found for node: %s", nodeID)
	}

	return repo, nil
}

// UpdateRepository updates a repository's information
func (s *RemoteService) UpdateRepository(ctx context.Context, nodeID string) error {
	s.repositoryMutex.Lock()
	defer s.repositoryMutex.Unlock()

	repo, exists := s.repositories[nodeID]
	if !exists {
		return fmt.Errorf("repository not found for node: %s", nodeID)
	}

	// Get latest repository information
	repository, _, err := s.githubClient.Repositories.Get(ctx, repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to get repository information: %v", err)
	}

	if repository.GetArchived() {
		repo.IsArchived = true
		return fmt.Errorf("repository has been archived")
	}

	// Update branch information
	branch, _, err := s.githubClient.Repositories.GetBranch(ctx, repo.Owner, repo.Name, repo.Branch, true)
	if err != nil {
		return fmt.Errorf("failed to get branch information: %v", err)
	}

	repo.LastCommit = branch.GetCommit().GetSHA()
	return nil
}

// CreatePullRequest creates a new pull request for changes
func (s *RemoteService) CreatePullRequest(ctx context.Context, nodeID string, title string, description string, changes []hephaestus.CodeChange) error {
	repo, err := s.GetRepository(ctx, nodeID)
	if err != nil {
		return err
	}

	if repo.IsArchived {
		return fmt.Errorf("cannot create pull request for archived repository")
	}

	// Create a new branch for the changes
	branchName := fmt.Sprintf("hephaestus/%s/%s", nodeID, repo.LastCommit[:8])
	ref := fmt.Sprintf("refs/heads/%s", branchName)

	// Create the branch from the latest commit
	refObj := &github.Reference{
		Ref: github.String(ref),
		Object: &github.GitObject{
			SHA: github.String(repo.LastCommit),
		},
	}

	_, _, err = s.githubClient.Git.CreateRef(ctx, repo.Owner, repo.Name, refObj)
	if err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	// Apply changes
	for _, change := range changes {
		// Get the current file content
		fileContent, _, _, err := s.githubClient.Repositories.GetContents(
			ctx,
			repo.Owner,
			repo.Name,
			change.FilePath,
			&github.RepositoryContentGetOptions{Ref: branchName},
		)
		if err != nil {
			return fmt.Errorf("failed to get file content: %v", err)
		}

		// Create the commit
		_, _, err = s.githubClient.Repositories.CreateFile(
			ctx,
			repo.Owner,
			repo.Name,
			change.FilePath,
			&github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("Update %s", change.FilePath)),
				Content: []byte(change.UpdatedCode),
				Branch:  github.String(branchName),
				SHA:     github.String(fileContent.GetSHA()),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to commit changes: %v", err)
		}
	}

	// Create pull request
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:               github.String(branchName),
		Base:               github.String(repo.Branch),
		Body:               github.String(description),
		MaintainerCanModify: github.Bool(true),
	}

	_, _, err = s.githubClient.PullRequests.Create(ctx, repo.Owner, repo.Name, newPR)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %v", err)
	}

	return nil
}

// Cleanup removes a repository instance and its resources
func (s *RemoteService) Cleanup(ctx context.Context, nodeID string) error {
	s.repositoryMutex.Lock()
	defer s.repositoryMutex.Unlock()

	if _, exists := s.repositories[nodeID]; !exists {
		return fmt.Errorf("repository not found for node: %s", nodeID)
	}

	delete(s.repositories, nodeID)
	return nil
} 
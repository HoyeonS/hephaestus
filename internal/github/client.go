package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"

	"github.com/HoyeonS/hephaestus/internal/repository"
)

// Client handles GitHub API operations
type Client struct {
	client *github.Client
	owner  string
	repo   string
}

// NewClient creates a new GitHub client
func NewClient(token, repoPath string) (*Client, error) {
	// Validate repository path
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository path: %s, expected format: owner/repo", repoPath)
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		owner:  parts[0],
		repo:   parts[1],
	}, nil
}

// FetchRepository fetches the entire repository and creates a virtual repository
func (c *Client) FetchRepository(ctx context.Context, branch string) (*repository.VirtualRepository, error) {
	// Get repository content
	_, directoryContent, _, err := c.client.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		"",
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository contents: %w", err)
	}

	// Create virtual repository
	vRepo := &repository.VirtualRepository{
		GitHubRepo: fmt.Sprintf("%s/%s", c.owner, c.repo),
		Branch:     branch,
		Files:      make(map[string]*repository.FileNode),
		LastSynced: time.Now(),
	}

	// Recursively fetch all files
	err = c.fetchDirectoryContents(ctx, "", directoryContent, vRepo)
	if err != nil {
		return nil, err
	}

	return vRepo, nil
}

// fetchDirectoryContents recursively fetches all files in a directory
func (c *Client) fetchDirectoryContents(ctx context.Context, path string, contents []*github.RepositoryContent, vRepo *repository.VirtualRepository) error {
	for _, content := range contents {
		if *content.Type == "dir" {
			// Fetch directory contents
			_, dirContents, _, err := c.client.Repositories.GetContents(
				ctx,
				c.owner,
				c.repo,
				*content.Path,
				&github.RepositoryContentGetOptions{},
			)
			if err != nil {
				return fmt.Errorf("failed to get directory contents for %s: %w", *content.Path, err)
			}

			err = c.fetchDirectoryContents(ctx, *content.Path, dirContents, vRepo)
			if err != nil {
				return err
			}
		} else {
			// Fetch file content
			fileContent, _, _, err := c.client.Repositories.GetContents(
				ctx,
				c.owner,
				c.repo,
				*content.Path,
				&github.RepositoryContentGetOptions{},
			)
			if err != nil {
				return fmt.Errorf("failed to get file contents for %s: %w", *content.Path, err)
			}

			content, err := fileContent.GetContent()
			if err != nil {
				return fmt.Errorf("failed to decode file contents for %s: %w", *content.Path, err)
			}

			// Create FileNode
			node := &repository.FileNode{
				ID:          generateFileID(*content.Path),
				Path:        *content.Path,
				Content:     content,
				Language:    detectLanguage(*content.Path),
				LastUpdated: time.Now(),
				Metadata:    generateMetadata(content, *content.Path),
			}

			vRepo.Files[*content.Path] = node
		}
	}

	return nil
}

// CreatePullRequest creates a new pull request with the specified changes
func (c *Client) CreatePullRequest(ctx context.Context, branch, title, body string, changes map[string]string) (*github.PullRequest, error) {
	// Create a new branch
	ref, _, err := c.client.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/main")
	if err != nil {
		return nil, fmt.Errorf("failed to get ref: %w", err)
	}

	newBranch := fmt.Sprintf("refs/heads/%s", branch)
	_, _, err = c.client.Git.CreateRef(ctx, c.owner, c.repo, &github.Reference{
		Ref:    &newBranch,
		Object: ref.Object,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Create commits for each file change
	for path, content := range changes {
		// Get the current file to update
		file, _, _, err := c.client.Repositories.GetContents(
			ctx,
			c.owner,
			c.repo,
			path,
			&github.RepositoryContentGetOptions{Ref: branch},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get file %s: %w", path, err)
		}

		// Create the commit
		_, _, err = c.client.Repositories.CreateFile(
			ctx,
			c.owner,
			c.repo,
			path,
			&github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("Update %s", path)),
				Content: []byte(content),
				Branch:  github.String(branch),
				SHA:     file.SHA,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create commit for %s: %w", path, err)
		}
	}

	// Create pull request
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:               github.String(branch),
		Base:               github.String("main"),
		Body:               github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := c.client.PullRequests.Create(ctx, c.owner, c.repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr, nil
}

// Helper functions

func generateFileID(path string) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), strings.ReplaceAll(path, "/", "-"))
}

func detectLanguage(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(path[strings.LastIndex(path, "."):], "."))
	switch ext {
	case "go":
		return "Go"
	case "js", "jsx":
		return "JavaScript"
	case "ts", "tsx":
		return "TypeScript"
	case "py":
		return "Python"
	case "java":
		return "Java"
	case "rb":
		return "Ruby"
	case "php":
		return "PHP"
	case "rs":
		return "Rust"
	default:
		return "Unknown"
	}
}

func generateMetadata(content *github.RepositoryContent, path string) repository.Metadata {
	return repository.Metadata{
		LineCount: strings.Count(*content.Content, "\n") + 1,
		GitInfo: repository.GitInfo{
			LastCommit:     *content.SHA,
			LastCommitDate: time.Now(), // This should be fetched from commit info
		},
	}
} 
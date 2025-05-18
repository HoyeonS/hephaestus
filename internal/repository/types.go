package repository

import (
	"time"
)

// FileNode represents a file in the virtual repository
type FileNode struct {
	ID          string    // Unique identifier for the file
	Path        string    // File path relative to repository root
	Content     string    // File content
	Language    string    // Programming language
	LastUpdated time.Time // Last update timestamp
	Metadata    Metadata  // Additional file metadata
}

// Metadata contains additional information about a file
type Metadata struct {
	LineCount    int      // Total number of lines
	Imports      []string // List of imports/dependencies
	GitInfo      GitInfo  // Git-related information
}

// GitInfo contains Git-related information about a file
type GitInfo struct {
	LastCommit     string
	LastAuthor     string
	LastCommitDate time.Time
	Branch         string
}

// VirtualRepository represents a collection of FileNodes
type VirtualRepository struct {
	ID            string               // Unique identifier for the repository
	GitHubRepo    string               // GitHub repository name (owner/repo)
	Branch        string               // Current branch
	Files         map[string]*FileNode // Map of file path to FileNode
	LastSynced    time.Time           // Last synchronization time
	Configuration *Configuration       // Repository configuration
}

// Configuration holds the repository configuration
type Configuration struct {
	GitHub GitHubConfig
	AI     AIConfig
	Log    LogConfig
	Mode   string // "suggest" or "deploy"
}

// GitHubConfig holds GitHub-related configuration
type GitHubConfig struct {
	Repository string
	Branch     string
	Token      string
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	Provider string
	APIKey   string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string // "debug", "info", "warn", "error", "fatal"
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewVirtualRepository creates a new virtual repository instance
func NewVirtualRepository(config *Configuration) (*VirtualRepository, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &VirtualRepository{
		ID:            generateID(),
		GitHubRepo:    config.GitHub.Repository,
		Branch:        config.GitHub.Branch,
		Files:         make(map[string]*FileNode),
		LastSynced:    time.Now(),
		Configuration: config,
	}, nil
}

// validateConfig validates the repository configuration
func validateConfig(config *Configuration) error {
	if config == nil {
		return &ValidationError{Field: "config", Message: "configuration cannot be nil"}
	}

	if config.GitHub.Repository == "" {
		return &ValidationError{Field: "github.repository", Message: "repository cannot be empty"}
	}

	if config.GitHub.Token == "" {
		return &ValidationError{Field: "github.token", Message: "GitHub token cannot be empty"}
	}

	if config.AI.Provider == "" {
		return &ValidationError{Field: "ai.provider", Message: "AI provider cannot be empty"}
	}

	if config.AI.APIKey == "" {
		return &ValidationError{Field: "ai.api_key", Message: "AI API key cannot be empty"}
	}

	if config.Log.Level == "" {
		return &ValidationError{Field: "log.level", Message: "log level cannot be empty"}
	}

	if config.Mode != "suggest" && config.Mode != "deploy" {
		return &ValidationError{Field: "mode", Message: "mode must be either 'suggest' or 'deploy'"}
	}

	return nil
}

// generateID generates a unique identifier for the repository
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
} 
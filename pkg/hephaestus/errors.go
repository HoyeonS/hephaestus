package hephaestus

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidConfig indicates an invalid configuration
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrNodeNotFound indicates a node was not found
	ErrNodeNotFound = errors.New("node not found")

	// ErrNodeAlreadyExists indicates a node already exists
	ErrNodeAlreadyExists = errors.New("node already exists")

	// ErrInvalidNodeStatus indicates an invalid node status
	ErrInvalidNodeStatus = errors.New("invalid node status")

	// ErrRepositoryNotInitialized indicates the repository is not initialized
	ErrRepositoryNotInitialized = errors.New("repository not initialized")

	// ErrFileNotFound indicates a file was not found
	ErrFileNotFound = errors.New("file not found")

	// ErrFileTooLarge indicates a file exceeds the size limit
	ErrFileTooLarge = errors.New("file too large")

	// ErrInvalidLogEntry indicates an invalid log entry
	ErrInvalidLogEntry = errors.New("invalid log entry")

	// ErrAIProviderError indicates an error from the AI provider
	ErrAIProviderError = errors.New("AI provider error")

	// ErrGitHubError indicates an error from GitHub
	ErrGitHubError = errors.New("GitHub error")

	// ErrInvalidSolution indicates an invalid solution
	ErrInvalidSolution = errors.New("invalid solution")

	// ErrOperationTimeout indicates an operation timed out
	ErrOperationTimeout = errors.New("operation timed out")
)

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	Field   string
	Message string
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s: %s", e.Field, e.Message)
}

// NodeError represents a node-related error
type NodeError struct {
	NodeID  string
	Message string
	Err     error
}

func (e *NodeError) Error() string {
	return fmt.Sprintf("node error (ID: %s): %s: %v", e.NodeID, e.Message, e.Err)
}

func (e *NodeError) Unwrap() error {
	return e.Err
}

// RepositoryError represents a repository-related error
type RepositoryError struct {
	Path    string
	Message string
	Err     error
}

func (e *RepositoryError) Error() string {
	return fmt.Sprintf("repository error (path: %s): %s: %v", e.Path, e.Message, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// LogProcessingError represents a log processing error
type LogProcessingError struct {
	NodeID  string
	Message string
	Err     error
}

func (e *LogProcessingError) Error() string {
	return fmt.Sprintf("log processing error (node: %s): %s: %v", e.NodeID, e.Message, e.Err)
}

func (e *LogProcessingError) Unwrap() error {
	return e.Err
}

// AIError represents an AI provider error
type AIError struct {
	Provider string
	Message  string
	Err      error
}

func (e *AIError) Error() string {
	return fmt.Sprintf("AI error (%s): %s: %v", e.Provider, e.Message, e.Err)
}

func (e *AIError) Unwrap() error {
	return e.Err
}

// GitHubError represents a GitHub API error
type GitHubError struct {
	Operation string
	Message   string
	Err       error
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("GitHub error (%s): %s: %v", e.Operation, e.Message, e.Err)
}

func (e *GitHubError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error
type ValidationError struct {
	Object  string
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error (%s.%s): %s", e.Object, e.Field, e.Message)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNodeNotFound) || errors.Is(err, ErrFileNotFound)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	var configValidationErr *ConfigValidationError
	return errors.As(err, &validationErr) || errors.As(err, &configValidationErr)
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrOperationTimeout)
}

// IsExternalError checks if an error is from an external service
func IsExternalError(err error) bool {
	var aiErr *AIError
	var githubErr *GitHubError
	return errors.As(err, &aiErr) || errors.As(err, &githubErr)
} 
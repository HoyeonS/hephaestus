package hephaestus

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrInvalidConfig indicates an invalid configuration
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrInvalidArgument indicates an invalid argument
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrUnavailable indicates a service is unavailable
	ErrUnavailable = errors.New("service unavailable")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")

	// ErrCanceled indicates an operation was canceled
	ErrCanceled = errors.New("operation canceled")

	// ErrModelProviderError indicates an error from the model provider
	ErrModelProviderError = errors.New("model provider error")

	// ErrRemoteRepositoryError indicates an error from the remote repository
	ErrRemoteRepositoryError = errors.New("remote repository error")

	// ErrNodeError indicates an error from a node
	ErrNodeError = errors.New("node error")

	// ErrLogError indicates a logging error
	ErrLogError = errors.New("log error")

	// ErrMetricsError indicates a metrics collection error
	ErrMetricsError = errors.New("metrics error")
)

// ModelError represents a model provider error
type ModelError struct {
	Provider string
	Message  string
	Err      error
}

func (e *ModelError) Error() string {
	return fmt.Sprintf("model error (%s): %s: %v", e.Provider, e.Message, e.Err)
}

func (e *ModelError) Unwrap() error {
	return e.Err
}

// RemoteRepositoryError represents a remote repository API error
type RemoteRepositoryError struct {
	Provider  string
	Operation string
	Message   string
	Err       error
}

func (e *RemoteRepositoryError) Error() string {
	return fmt.Sprintf("remote repository error (%s): %s: %v", e.Operation, e.Message, e.Err)
}

func (e *RemoteRepositoryError) Unwrap() error {
	return e.Err
}

// NodeError represents a node-specific error
type NodeError struct {
	NodeID  string
	Message string
	Err     error
}

func (e *NodeError) Error() string {
	return fmt.Sprintf("node error (%s): %s: %v", e.NodeID, e.Message, e.Err)
}

func (e *NodeError) Unwrap() error {
	return e.Err
}

// IsProviderError checks if the error is from an external provider
func IsProviderError(err error) bool {
	var modelErr *ModelError
	var repoErr *RemoteRepositoryError
	return errors.As(err, &modelErr) || errors.As(err, &repoErr)
}

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	Field   string
	Message string
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s: %s", e.Field, e.Message)
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
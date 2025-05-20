package hephaestus

import (
	"fmt"
)

// ClientConfiguration represents the client-side configuration
type ClientConfiguration struct {
	// Remote Repository Settings
	RemoteRepo RemoteRepoConfiguration `json:"remote_repo" yaml:"remote_repo"`

	// Logging Settings
	Logging LoggingConfiguration `json:"logging" yaml:"logging"`

	// Repository Settings
	Repository RepositoryConfiguration `json:"repository" yaml:"repository"`

	// Base URL for API endpoints
	BaseURL string `json:"base_url" yaml:"base_url"`
}

// LoggingConfiguration contains logging settings
type LoggingConfiguration struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// RepositoryConfiguration contains repository settings
type RepositoryConfiguration struct {
	Path         string `json:"path" yaml:"path"`
	FileLimit    int    `json:"file_limit" yaml:"file_limit"`
	FileSizeLimit int64  `json:"file_size_limit" yaml:"file_size_limit"`
}

// ValidateClientConfiguration validates the client configuration
func ValidateClientConfiguration(config *ClientConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate logging settings
	if config.Logging.Level == "" {
		return fmt.Errorf("logging level is required")
	}
	if !isValidLogLevel(config.Logging.Level) {
		return fmt.Errorf("invalid logging level")
	}
	if config.Logging.Format == "" {
		return fmt.Errorf("logging format is required")
	}
	if !isValidLogFormat(config.Logging.Format) {
		return fmt.Errorf("invalid logging format")
	}

	// Validate repository settings
	if config.Repository.Path == "" {
		return fmt.Errorf("repository path is required")
	}
	if config.Repository.FileLimit <= 0 {
		return fmt.Errorf("file limit must be positive")
	}
	if config.Repository.FileSizeLimit <= 0 {
		return fmt.Errorf("file size limit must be positive")
	}

	// Validate base URL
	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	return nil
}

// isValidLogFormat checks if the log format is valid
func isValidLogFormat(format string) bool {
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	return validFormats[format]
} 
package hephaestus

import (
	"fmt"
	"time"
)

// SystemConfiguration represents the native Hephaestus configuration
type SystemConfiguration struct {
	// Model Settings
	ModelConfiguration ModelConfiguration `json:"model" yaml:"model"`

	// Limit Settings
	LimitConfiguration LimitConfiguration `json:"limit" yaml:"limit"`
}

// ModelConfiguration contains model settings
type ModelConfiguration struct {
	ModelServiceProvider string `json:"service_provider" yaml:"service_provider"`
	ModelServiceAPIKey   string `json:"service_api_key" yaml:"service_api_key"`
	ModelVersion         string `json:"model_version" yaml:"model_version"`
}

// RepositoryConfiguration contains repository settings
type LimitConfiguration struct {
	LogChunkLimit      int `json:"log_chunk_limit" yaml:"log_chunk_limit"`
	FileNodeCountLimit int `json:"file_node_count_limit" yaml:"file_node_count_limit"`
}

// ClientConfiguration represents the client side Hephaestus Node Level configuration
type ClientNodeConfiguration struct {
	// Log Processing Settings
	LogProcessingConfiguration LogProcessingConfiguration `json:"log" yaml:"log"`

	// Remote Repository Settings
	RemoteRepositoryConfiguration RemoteRepositoryConfiguration `json:"remote-repository" yaml:"remote-repository"`
}

// LogProcessingConfiguration contains log processing settings
type LogProcessingConfiguration struct {
	ThresholdLevel string `json:"threshold_level" yaml:"threshold_level"`
}

// Remote Repository Provider contains remote repository code base connection settings
type RemoteRepositoryConfiguration struct {
	RemoteRepositoryProvider string `json:"remote_repository_provider" yaml: "remote_repository_provider"`
	RepositoryAddress        string `json:"address" yaml:"address"`
	BaseDirectory            string `json:"base_dir" yaml:"base_dir"`
	ProviderToken            string `json:"provider_token" yaml:"provider_token"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context"`
	ErrorTrace  string                 `json:"error_trace,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// Solution represents a generated solution
type Solution struct {
	ID          string    `json:"id"`
	LogEntry    LogEntry  `json:"log_entry"`
	Description string    `json:"description"`
	CodeChanges []Change  `json:"code_changes"`
	GeneratedAt time.Time `json:"generated_at"`
	Confidence  float64   `json:"confidence"`
}

// Change represents a code change
type Change struct {
	FilePath    string `json:"file_path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	OldContent  string `json:"old_content"`
	NewContent  string `json:"new_content"`
	Description string `json:"description"`
}

// NodeStatus represents the node status
type NodeStatus string

const (
	NodeStatusInitializing NodeStatus = "initializing"
	NodeStatusOperational  NodeStatus = "operational"
	NodeStatusProcessing   NodeStatus = "processing"
	NodeStatusError        NodeStatus = "error"
)

// ConfigurationValidationError represents a configuration validation error
type ConfigurationValidationError struct {
	FieldName    string
	ErrorMessage string
}

func (e ConfigurationValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.FieldName, e.ErrorMessage)
}

// ValidateSystemConfiguration validates the system configuration
func ValidateSystemConfiguration(config *SystemConfiguration) error {
	if config == nil {
		return &ConfigurationValidationError{FieldName: "config", ErrorMessage: "configuration cannot be nil"}
	}

	return nil
}

func ValidateClientNodeConfiguration(config *ClientNodeConfiguration) error {
	if config == nil {
		return &ConfigurationValidationError{FieldName: "config", ErrorMessage: "configuration cannot be nil"}
	}

	return nil
}

package hephaestus

import (
	"context"
	"fmt"
	"time"
)

// SystemNode represents a Hephaestus system node instance
type SystemNode struct {
	NodeIdentifier string
	CurrentStatus  NodeOperationalStatus
	NodeConfig     *SystemConfiguration
	CreationTime   time.Time
	LastUpdateTime time.Time
}

// NodeOperationalStatus represents the current operational state of a node
type NodeOperationalStatus string

const (
	NodeStatusInitializing NodeOperationalStatus = "initializing"
	NodeStatusOperational  NodeOperationalStatus = "operational"
	NodeStatusProcessing   NodeOperationalStatus = "processing"
	NodeStatusInactive     NodeOperationalStatus = "inactive"
	NodeStatusError       NodeOperationalStatus = "error"
)

// SystemConfiguration represents the complete configuration for a Hephaestus node
type SystemConfiguration struct {
	RemoteSettings    RemoteRepositoryConfiguration `json:"remote" yaml:"remote"`
	ModelSettings     ModelServiceConfiguration     `json:"model" yaml:"model"`
	LoggingSettings   LoggingConfiguration         `json:"log" yaml:"log"`
	OperationalMode   string                       `json:"mode" yaml:"mode"`
	RepositorySettings RepositoryConfiguration     `json:"repository" yaml:"repository"`
}

// RemoteRepositoryConfiguration contains remote repository integration settings
type RemoteRepositoryConfiguration struct {
	AuthToken        string `json:"token" yaml:"token"`
	RepositoryOwner  string `json:"owner" yaml:"owner"`
	RepositoryName   string `json:"repository" yaml:"repository"`
	TargetBranch     string `json:"branch" yaml:"branch"`
}

// ModelServiceConfiguration contains model service provider settings
type ModelServiceConfiguration struct {
	ServiceProvider string `json:"provider" yaml:"provider"`
	ServiceAPIKey   string `json:"api_key" yaml:"api_key"`
	ModelVersion    string `json:"model_version" yaml:"model_version"`
}

// LoggingConfiguration contains logging system settings
type LoggingConfiguration struct {
	LogLevel     string `json:"level" yaml:"level"`
	OutputFormat string `json:"format" yaml:"format"`
}

// RepositoryConfiguration contains repository management settings
type RepositoryConfiguration struct {
	RepositoryPath string `json:"path" yaml:"path"`
	FileLimit      int    `json:"max_files" yaml:"max_files"`
	FileSizeLimit  int64  `json:"max_file_size" yaml:"max_file_size"`
}

// LogEntryData represents a structured log entry from the client
type LogEntryData struct {
	NodeIdentifier string            `json:"node_id"`
	LogLevel       string            `json:"level"`
	LogMessage     string            `json:"message"`
	LogTimestamp   time.Time         `json:"timestamp"`
	LogMetadata    map[string]string `json:"metadata"`
	ErrorTrace     string            `json:"stack_trace"`
}

// ProposedSolution represents a generated fix proposal for an issue
type ProposedSolution struct {
	SolutionID       string        `json:"id"`
	NodeIdentifier   string        `json:"node_id"`
	AssociatedLog    *LogEntryData `json:"log_entry"`
	ProposedChanges  string        `json:"suggestion"`
	AffectedFiles    []string      `json:"files"`
	GenerationTime   time.Time     `json:"created_at"`
	ConfidenceScore  float64       `json:"confidence"`
}

// NodeLifecycleManager manages the complete lifecycle of system nodes
type NodeLifecycleManager interface {
	// CreateSystemNode initializes a new node with the provided configuration
	CreateSystemNode(ctx context.Context, config *SystemConfiguration) (*SystemNode, error)
	
	// GetSystemNode retrieves node information by identifier
	GetSystemNode(ctx context.Context, nodeIdentifier string) (*SystemNode, error)
	
	// DeleteSystemNode removes a node from the system
	DeleteSystemNode(ctx context.Context, nodeIdentifier string) error
	
	// UpdateNodeOperationalStatus updates the operational status of a node
	UpdateNodeOperationalStatus(ctx context.Context, nodeIdentifier string, status NodeOperationalStatus) error
}

// LogProcessingService handles log entry processing and analysis
type LogProcessingService interface {
	// ProcessLogEntry processes a single log entry
	ProcessLogEntry(ctx context.Context, entry *LogEntryData) error
	
	// StreamLogEntries initiates streaming of logs for a specific node
	StreamLogEntries(ctx context.Context, nodeIdentifier string) (<-chan *LogEntryData, <-chan error)
}

// RepositoryManager manages the virtual repository system
type RepositoryManager interface {
	// InitializeRepository sets up the repository environment
	InitializeRepository(ctx context.Context, config *RepositoryConfiguration) error
	
	// GetFileContents retrieves contents of a specific file
	GetFileContents(ctx context.Context, filePath string) ([]byte, error)
	
	// UpdateFileContents updates the contents of a specific file
	UpdateFileContents(ctx context.Context, filePath string, contents []byte) error
	
	// ListRepositoryFiles lists all files in the repository
	ListRepositoryFiles(ctx context.Context) ([]string, error)
}

// ModelServiceProvider provides model-based analysis and solution generation
type ModelServiceProvider interface {
	// GenerateSolutionProposal generates a solution for a given log entry
	GenerateSolutionProposal(ctx context.Context, entry *LogEntryData, repo RepositoryManager) (*ProposedSolution, error)
	
	// ValidateSolutionProposal validates a generated solution
	ValidateSolutionProposal(ctx context.Context, solution *ProposedSolution) error
}

// RemoteRepositoryService manages remote repository operations
type RemoteRepositoryService interface {
	// CreateChangeRequest creates a change request for a proposed solution
	CreateChangeRequest(ctx context.Context, solution *ProposedSolution) (string, error)
	
	// SynchronizeRepository ensures local and remote repositories are in sync
	SynchronizeRepository(ctx context.Context) error
}

// MetricsCollectionService collects and manages system metrics
type MetricsCollectionService interface {
	// RecordOperationMetrics records metrics for an operation
	RecordOperationMetrics(operationName string, duration time.Duration, successful bool)
	
	// RecordErrorMetrics records error-related metrics
	RecordErrorMetrics(componentName string, err error)
	
	// GetCurrentMetrics retrieves current system metrics
	GetCurrentMetrics() map[string]interface{}
}

// RepositoryFileNode represents a file in the virtual repository system
type RepositoryFileNode struct {
	FileIdentifier string            `json:"id"`
	FilePath       string            `json:"path"`     // relative to repo root
	FileContents   string            `json:"content"`
	FileLanguage   string            `json:"language"` // programming language
	LastModified   time.Time         `json:"last_updated"`
	FileMetadata   FileMetadataInfo  `json:"metadata"`
}

// FileMetadataInfo contains extended file information
type FileMetadataInfo struct {
	TotalLines    int      `json:"line_count"`
	Dependencies  []string `json:"imports"`      // list of imports/dependencies
	VersionInfo   VersionControlInfo  `json:"version_info"`
}

// VersionControlInfo contains version control specific file information
type VersionControlInfo struct {
	LastCommitHash string    `json:"last_commit"`
	LastCommitter  string    `json:"last_author"`
	CommitTime     time.Time `json:"last_commit_date"`
}

// VirtualRepositorySystem represents the complete virtual repository
type VirtualRepositorySystem struct {
	RepositoryID    string                    `json:"id"`
	RemoteReference string                    `json:"remote_repo"`
	TargetBranch    string                    `json:"branch"`
	FileCollection  map[string]*RepositoryFileNode `json:"files"`
	LastSyncTime    time.Time                 `json:"last_synced"`
	SystemConfig    *SystemConfiguration      `json:"configuration"`
}

// InitializationResponse represents the response from node initialization
type InitializationResponse struct {
	OperationalStatus string `json:"status"`
	StatusMessage     string `json:"message"`
	NodeIdentifier    string `json:"node_id"`
}

// CodeModification represents a specific code change in a solution
type CodeModification struct {
	TargetFile    string `json:"file_path"`
	OriginalCode  string `json:"original_code"`
	ModifiedCode  string `json:"updated_code"`
	StartLine     int    `json:"line_start"`
	EndLine       int    `json:"line_end"`
}

// ConfigurationValidationError represents a configuration validation error
type ConfigurationValidationError struct {
	FieldName    string
	ErrorMessage string
}

func (e ConfigurationValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.FieldName, e.ErrorMessage)
}

// ValidateSystemConfiguration performs comprehensive validation of the system configuration
func ValidateSystemConfiguration(config *SystemConfiguration) error {
	if config == nil {
		return &ConfigurationValidationError{FieldName: "config", ErrorMessage: "configuration cannot be nil"}
	}

	// Validate Remote Repository configuration
	if config.RemoteSettings.RepositoryName == "" {
		return &ConfigurationValidationError{FieldName: "remote.repository", ErrorMessage: "repository name cannot be empty"}
	}
	if config.RemoteSettings.AuthToken == "" {
		return &ConfigurationValidationError{FieldName: "remote.token", ErrorMessage: "authentication token cannot be empty"}
	}
	if config.RemoteSettings.TargetBranch == "" {
		config.RemoteSettings.TargetBranch = "main" // Set default branch
	}

	// Validate Model configuration
	if config.ModelSettings.ServiceProvider == "" {
		return &ConfigurationValidationError{FieldName: "model.provider", ErrorMessage: "model service provider cannot be empty"}
	}
	if config.ModelSettings.ServiceAPIKey == "" {
		return &ConfigurationValidationError{FieldName: "model.api_key", ErrorMessage: "model service API key cannot be empty"}
	}

	// Validate Log configuration
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[config.LoggingSettings.LogLevel] {
		return &ConfigurationValidationError{FieldName: "log.level", ErrorMessage: "invalid log level specified"}
	}

	// Validate Mode
	validModes := map[string]bool{"suggest": true, "deploy": true}
	if !validModes[config.OperationalMode] {
		return &ConfigurationValidationError{FieldName: "mode", ErrorMessage: "operational mode must be either 'suggest' or 'deploy'"}
	}

	return nil
} 
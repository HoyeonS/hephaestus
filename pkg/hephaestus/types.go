package hephaestus

import (
	"context"
	"fmt"
	"time"
)

// Config represents the configuration for the Hephaestus service
type Config struct {
	Repository *RepositoryConfig `yaml:"repository"`
	Model      *ModelConfig      `yaml:"model"`
	Log        *LogConfig        `yaml:"log"`
}

// RepositoryConfig represents the configuration for the repository service
type RepositoryConfig struct {
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
	Token    string `yaml:"token"`
	Branch   string `yaml:"branch"`
	BasePath string `yaml:"base_path"`
}

// ModelConfig represents the configuration for the model service
type ModelConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

// LogConfig represents the configuration for logging
type LogConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

// NodeStatus represents the operational status of a node
type NodeStatus string

// SystemNode represents a node in the system
type SystemNode struct {
	ID         string     `json:"id"`
	Status     NodeStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastActive time.Time  `json:"last_active"`
}

// LogEntryData represents a log entry from a node
type LogEntryData struct {
	NodeIdentifier string    `json:"node_identifier"`
	LogMessage     string    `json:"log_message"`
	LogLevel       string    `json:"log_level"`
	ErrorTrace     string    `json:"error_trace,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// ProposedSolution represents a solution proposed by the model
type ProposedSolution struct {
	SolutionID      string        `json:"solution_id"`
	NodeIdentifier  string        `json:"node_identifier"`
	AssociatedLog   *LogEntryData `json:"associated_log"`
	ProposedChanges string        `json:"proposed_changes"`
	GenerationTime  time.Time     `json:"generation_time"`
	ConfidenceScore float64       `json:"confidence_score"`
}

// Issue represents a problem identified in the system
type Issue struct {
	IssueID     string    `json:"issue_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

// CodeChange represents a change to be made to the codebase
type CodeChange struct {
	FilePath    string `json:"file_path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	OldContent  string `json:"old_content"`
	NewContent  string `json:"new_content"`
	Description string `json:"description"`
}

// MetricsCollectionService defines the interface for collecting metrics
type MetricsCollectionService interface {
	RecordNodeStatus(nodeID string, status string)
	RecordOperationMetrics(operation string, duration time.Duration, success bool)
	RecordErrorMetrics(source string, err error)
}

// RemoteRepositoryService defines the interface for interacting with remote repositories
type RemoteRepositoryService interface {
	Initialize(ctx context.Context) error
	GetFileContents(ctx context.Context, path string) ([]byte, error)
	UpdateFileContents(ctx context.Context, path string, content []byte, message string) error
	CreatePullRequest(ctx context.Context, title, body, head, base string) (string, error)
}

// RepositoryManager defines the interface for managing repositories
type RepositoryManager interface {
	Initialize(ctx context.Context) error
	GetFileContents(ctx context.Context, path string) ([]byte, error)
	UpdateFileContents(ctx context.Context, path string, content []byte, message string) error
	CreatePullRequest(ctx context.Context, title, body, head, base string) (string, error)
}

// NodeManager defines the interface for managing nodes
type NodeManager interface {
	RegisterNode(ctx context.Context, nodeID string, logLevel string, logOutput string) error
	GetNode(ctx context.Context, nodeID string) (*SystemNode, error)
	UpdateNodeStatus(ctx context.Context, nodeID string, status NodeStatus) error
	RemoveNode(ctx context.Context, nodeID string) error
}

// ModelService defines the interface for model operations
type ModelService interface {
	GenerateSolution(ctx context.Context, prompt string) (string, error)
	ValidateSolution(ctx context.Context, solution string) (bool, error)
}

// RepositoryService defines the interface for repository operations
type RepositoryService interface {
	Initialize(ctx context.Context) error
	ApplyChanges(ctx context.Context, changes string) error
	ValidateChanges(ctx context.Context, changes string) error
	Cleanup(ctx context.Context) error
}

// RemoteService defines the interface for remote operations
type RemoteService interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	SendMessage(ctx context.Context, message string) error
	ReceiveMessage(ctx context.Context) (string, error)
}

// NodeOperationalStatus represents the current operational state of a node
type NodeOperationalStatus string

const (
	NodeStatusInitializing NodeOperationalStatus = "initializing"
	NodeStatusOperational  NodeOperationalStatus = "operational"
	NodeStatusProcessing   NodeOperationalStatus = "processing"
	NodeStatusInactive     NodeOperationalStatus = "inactive"
	NodeStatusError        NodeOperationalStatus = "error"
)

// SystemConfiguration represents the complete configuration for a Hephaestus node
type SystemConfiguration struct {
	RemoteSettings     RemoteRepositoryConfiguration `json:"remote" yaml:"remote"`
	ModelSettings      ModelServiceConfiguration     `json:"model" yaml:"model"`
	LoggingSettings    LoggingConfiguration          `json:"log" yaml:"log"`
	OperationalMode    string                        `json:"mode" yaml:"mode"`
	RepositorySettings RepositoryConfiguration       `json:"repository" yaml:"repository"`
}

// RemoteRepositoryConfiguration contains remote repository integration settings
type RemoteRepositoryConfiguration struct {
	AuthToken       string `json:"token" yaml:"token"`
	RepositoryOwner string `json:"owner" yaml:"owner"`
	RepositoryName  string `json:"repository" yaml:"repository"`
	TargetBranch    string `json:"branch" yaml:"branch"`
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

// RepositoryFileNode represents a file in the virtual repository system
type RepositoryFileNode struct {
	FileIdentifier string           `json:"id"`
	FilePath       string           `json:"path"` // relative to repo root
	FileContents   string           `json:"content"`
	FileLanguage   string           `json:"language"` // programming language
	LastModified   time.Time        `json:"last_updated"`
	FileMetadata   FileMetadataInfo `json:"metadata"`
}

// FileMetadataInfo contains extended file information
type FileMetadataInfo struct {
	TotalLines   int                `json:"line_count"`
	Dependencies []string           `json:"imports"` // list of imports/dependencies
	VersionInfo  VersionControlInfo `json:"version_info"`
}

// VersionControlInfo contains version control specific file information
type VersionControlInfo struct {
	LastCommitHash string    `json:"last_commit"`
	LastCommitter  string    `json:"last_author"`
	CommitTime     time.Time `json:"last_commit_date"`
}

// VirtualRepositorySystem represents the complete virtual repository
type VirtualRepositorySystem struct {
	RepositoryID    string                         `json:"id"`
	RemoteReference string                         `json:"remote_repo"`
	TargetBranch    string                         `json:"branch"`
	FileCollection  map[string]*RepositoryFileNode `json:"files"`
	LastSyncTime    time.Time                      `json:"last_synced"`
	SystemConfig    *SystemConfiguration           `json:"configuration"`
}

// InitializationResponse represents the response from node initialization
type InitializationResponse struct {
	OperationalStatus string `json:"status"`
	StatusMessage     string `json:"message"`
	NodeIdentifier    string `json:"node_id"`
}

// CodeModification represents a specific code change in a solution
type CodeModification struct {
	TargetFile   string `json:"file_path"`
	OriginalCode string `json:"original_code"`
	ModifiedCode string `json:"updated_code"`
	StartLine    int    `json:"line_start"`
	EndLine      int    `json:"line_end"`
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

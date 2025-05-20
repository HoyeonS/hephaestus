package hephaestus

import (
	"fmt"
	"time"
)

// SystemConfiguration represents the native Hephaestus configuration
type SystemConfiguration struct {
	// Model Settings
	Model ModelConfiguration `json:"model" yaml:"model"`

	// Log Processing Settings
	LogSettings LogProcessingConfiguration `json:"log" yaml:"log"`

	// Operation Mode
	OperationalMode string `json:"mode" yaml:"mode"`
}

// ModelConfiguration contains model settings
type ModelConfiguration struct {
	Provider    string `json:"provider" yaml:"provider"`
	APIKey      string `json:"api_key" yaml:"api_key"`
	ModelVersion string `json:"model_version" yaml:"model_version"`
}

// LogProcessingConfiguration contains log processing settings
type LogProcessingConfiguration struct {
	ThresholdLevel  string        `json:"threshold_level" yaml:"threshold_level"`
	ThresholdCount  int           `json:"threshold_count" yaml:"threshold_count"`
	ThresholdWindow time.Duration `json:"threshold_window" yaml:"threshold_window"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context"`
	ErrorTrace  string                 `json:"error_trace,omitempty"`
	ProcessedAt time.Time             `json:"processed_at"`
}

// Solution represents a generated solution
type Solution struct {
	ID            string    `json:"id"`
	LogEntry      LogEntry  `json:"log_entry"`
	Description   string    `json:"description"`
	CodeChanges   []Change  `json:"code_changes"`
	GeneratedAt   time.Time `json:"generated_at"`
	Confidence    float64   `json:"confidence"`
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

	// Validate model settings
	if config.Model.Provider == "" {
		return &ConfigurationValidationError{FieldName: "model.provider", ErrorMessage: "model provider is required"}
	}
	if config.Model.APIKey == "" {
		return &ConfigurationValidationError{FieldName: "model.api_key", ErrorMessage: "model API key is required"}
	}
	if config.Model.ModelVersion == "" {
		return &ConfigurationValidationError{FieldName: "model.model_version", ErrorMessage: "model version is required"}
	}

	// Validate log settings
	if config.LogSettings.ThresholdLevel == "" {
		return &ConfigurationValidationError{FieldName: "log.threshold_level", ErrorMessage: "log threshold level is required"}
	}
	if !isValidLogLevel(config.LogSettings.ThresholdLevel) {
		return &ConfigurationValidationError{FieldName: "log.threshold_level", ErrorMessage: "invalid log threshold level"}
	}
	if config.LogSettings.ThresholdCount <= 0 {
		return &ConfigurationValidationError{FieldName: "log.threshold_count", ErrorMessage: "threshold count must be positive"}
	}
	if config.LogSettings.ThresholdWindow <= 0 {
		return &ConfigurationValidationError{FieldName: "log.threshold_window", ErrorMessage: "threshold window must be positive"}
	}

	// Validate mode
	if !isValidMode(config.OperationalMode) {
		return &ConfigurationValidationError{FieldName: "mode", ErrorMessage: "invalid operational mode"}
	}

	return nil
}

// isValidLogLevel checks if the log level is valid
func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	return validLevels[level]
}

// isValidMode checks if the mode is valid
func isValidMode(mode string) bool {
	validModes := map[string]bool{
		"suggest": true,
		"deploy":  true,
	}
	return validModes[mode]
}
 
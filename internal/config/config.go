package config

import (
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// ConfigurationManager handles configuration lifecycle and validation
type ConfigurationManager struct {
	currentConfig *hephaestus.SystemConfiguration
}

// NewConfigurationManager creates a new configuration manager instance
func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{}
}

// LoadConfiguration loads configuration from all available sources and validates it
func (cm *ConfigurationManager) LoadConfiguration(configPath string) error {
	// Load from file
	if err := cm.loadConfigurationFromFile(configPath); err != nil {
		return err
	}

	// Load from environment
	if err := cm.loadConfigurationFromEnvironment(); err != nil {
		return err
	}

	// Validate configuration
	if err := cm.validateConfiguration(); err != nil {
		return err
	}

	return nil
}

// GetCurrentConfiguration returns the current validated configuration
func (cm *ConfigurationManager) GetCurrentConfiguration() *hephaestus.SystemConfiguration {
	return cm.currentConfig
}

// UpdateConfiguration sets and validates a new configuration
func (cm *ConfigurationManager) UpdateConfiguration(config *hephaestus.SystemConfiguration) error {
	if err := cm.validateConfigurationObject(config); err != nil {
		return err
	}
	cm.currentConfig = config
	return nil
}

// validateConfiguration validates the current configuration
func (cm *ConfigurationManager) validateConfiguration() error {
	return cm.validateConfigurationObject(cm.currentConfig)
}

// validateConfigurationObject performs comprehensive validation of a configuration object
func (cm *ConfigurationManager) validateConfigurationObject(config *hephaestus.SystemConfiguration) error {
	if config == nil {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "config",
			ErrorMessage: "configuration is nil",
		}
	}

	// Validate Remote Repository configuration
	if err := cm.validateRemoteRepositoryConfiguration(&config.RemoteSettings); err != nil {
		return err
	}

	// Validate Model configuration
	if err := cm.validateModelConfiguration(&config.ModelSettings); err != nil {
		return err
	}

	// Validate Log configuration
	if err := cm.validateLoggingConfiguration(&config.LoggingSettings); err != nil {
		return err
	}

	// Validate Repository configuration
	if err := cm.validateRepositoryConfiguration(&config.RepositorySettings); err != nil {
		return err
	}

	// Validate Mode
	if err := cm.validateOperationMode(config.OperationalMode); err != nil {
		return err
	}

	return nil
}

// validateRemoteRepositoryConfiguration validates remote repository configuration settings
func (cm *ConfigurationManager) validateRemoteRepositoryConfiguration(config *hephaestus.RemoteRepositoryConfiguration) error {
	if config.AuthToken == "" {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "remote.token",
			ErrorMessage: "remote repository authentication token is required",
		}
	}

	if config.RepositoryOwner == "" {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "remote.owner",
			ErrorMessage: "remote repository owner is required",
		}
	}

	if config.RepositoryName == "" {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "remote.repository",
			ErrorMessage: "remote repository name is required",
		}
	}

	if config.TargetBranch == "" {
		config.TargetBranch = "main" // Set default branch
	}

	return nil
}

// validateModelConfiguration validates model service configuration settings
func (cm *ConfigurationManager) validateModelConfiguration(config *hephaestus.ModelServiceConfiguration) error {
	if config.ServiceProvider == "" {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "model.provider",
			ErrorMessage: "model service provider is required",
		}
	}

	if config.ServiceAPIKey == "" {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "model.api_key",
			ErrorMessage: "model service API key is required",
		}
	}

	return nil
}

// validateLoggingConfiguration validates logging configuration settings
func (cm *ConfigurationManager) validateLoggingConfiguration(config *hephaestus.LoggingConfiguration) error {
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[config.LogLevel] {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "log.level",
			ErrorMessage: "invalid logging level specified",
		}
	}

	validLogFormats := map[string]bool{
		"json":  true,
		"text": true,
	}

	if !validLogFormats[config.OutputFormat] {
		config.OutputFormat = "json" // Set default format
	}

	return nil
}

// validateRepositoryConfiguration validates repository configuration settings
func (cm *ConfigurationManager) validateRepositoryConfiguration(config *hephaestus.RepositoryConfiguration) error {
	if config.RepositoryPath == "" {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "repository.path",
			ErrorMessage: "repository filesystem path is required",
		}
	}

	if config.FileLimit <= 0 {
		config.FileLimit = 10000 // Set default max files
	}

	if config.FileSizeLimit <= 0 {
		config.FileSizeLimit = 1 << 20 // Set default max file size (1MB)
	}

	return nil
}

// validateOperationMode validates the system operation mode
func (cm *ConfigurationManager) validateOperationMode(mode string) error {
	validOperationModes := map[string]bool{
		"suggest": true,
		"deploy":  true,
	}

	if !validOperationModes[mode] {
		return &hephaestus.ConfigurationValidationError{
			FieldName:    "mode",
			ErrorMessage: "invalid operation mode specified",
		}
	}

	return nil
} 
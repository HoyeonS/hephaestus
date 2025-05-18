package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"gopkg.in/yaml.v2"
)

// ConfigurationManager handles configuration lifecycle and validation
type ConfigurationManager struct {
	currentConfig *hephaestus.SystemConfiguration
}

// NewConfigurationManager creates a new configuration manager instance
func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{}
}

// Get returns the current configuration
func (cm *ConfigurationManager) Get() *hephaestus.SystemConfiguration {
	return cm.currentConfig
}

// Set validates and sets a new configuration
func (cm *ConfigurationManager) Set(config *hephaestus.SystemConfiguration) error {
	if err := validateSystemConfiguration(config); err != nil {
		return err
	}
	cm.currentConfig = config
	return nil
}

// SaveConfigToFile saves the configuration to a file
func SaveConfigToFile(config *hephaestus.SystemConfiguration, filePath string) error {
	if err := validateSystemConfiguration(config); err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// LoadConfigFromFile loads configuration from a file
func LoadConfigFromFile(filePath string) (*hephaestus.SystemConfiguration, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	config := &hephaestus.SystemConfiguration{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	if err := validateSystemConfiguration(config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfigFromEnvironment loads configuration from environment variables
func LoadConfigFromEnvironment() *hephaestus.SystemConfiguration {
	config := &hephaestus.SystemConfiguration{
		RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
			AuthToken:       os.Getenv("HEPHAESTUS_REMOTE_TOKEN"),
			RepositoryOwner: os.Getenv("HEPHAESTUS_REMOTE_OWNER"),
			RepositoryName:  os.Getenv("HEPHAESTUS_REMOTE_REPO"),
			TargetBranch:   os.Getenv("HEPHAESTUS_REMOTE_BRANCH"),
		},
		ModelSettings: hephaestus.ModelServiceConfiguration{
			ServiceProvider: os.Getenv("HEPHAESTUS_MODEL_PROVIDER"),
			ServiceAPIKey:   os.Getenv("HEPHAESTUS_MODEL_API_KEY"),
			ModelVersion:    os.Getenv("HEPHAESTUS_MODEL_VERSION"),
		},
		LoggingSettings: hephaestus.LoggingConfiguration{
			LogLevel:     os.Getenv("HEPHAESTUS_LOG_LEVEL"),
			OutputFormat: os.Getenv("HEPHAESTUS_LOG_FORMAT"),
		},
		RepositorySettings: hephaestus.RepositoryConfiguration{
			RepositoryPath: os.Getenv("HEPHAESTUS_REPO_PATH"),
		},
		OperationalMode: os.Getenv("HEPHAESTUS_MODE"),
	}

	// Set defaults if not provided
	if config.LoggingSettings.LogLevel == "" {
		config.LoggingSettings.LogLevel = "info"
	}
	if config.LoggingSettings.OutputFormat == "" {
		config.LoggingSettings.OutputFormat = "json"
	}
	if config.OperationalMode == "" {
		config.OperationalMode = "suggest"
	}
	if config.RepositorySettings.FileLimit == 0 {
		config.RepositorySettings.FileLimit = 10000
	}
	if config.RepositorySettings.FileSizeLimit == 0 {
		config.RepositorySettings.FileSizeLimit = 1 << 20
	}

	return config
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *hephaestus.SystemConfiguration {
	return &hephaestus.SystemConfiguration{
		LoggingSettings: hephaestus.LoggingConfiguration{
			LogLevel:     "info",
			OutputFormat: "json",
		},
		RepositorySettings: hephaestus.RepositoryConfiguration{
			FileLimit:     10000,
			FileSizeLimit: 1 << 20,
		},
		OperationalMode: "suggest",
	}
}

// GetConfigFilePath returns the path to the configuration file
func GetConfigFilePath() string {
	if path := os.Getenv("HEPHAESTUS_CONFIG"); path != "" {
		return path
	}
	return filepath.Join(os.Getenv("HOME"), ".hephaestus", "config.yaml")
}

// validateSystemConfiguration validates the configuration
func validateSystemConfiguration(config *hephaestus.SystemConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate remote settings
	if config.RemoteSettings.AuthToken == "" {
		return fmt.Errorf("remote auth token is required")
	}
	if config.RemoteSettings.RepositoryOwner == "" {
		return fmt.Errorf("remote repository owner is required")
	}
	if config.RemoteSettings.RepositoryName == "" {
		return fmt.Errorf("remote repository name is required")
	}

	// Validate model settings
	if config.ModelSettings.ServiceProvider == "" {
		return fmt.Errorf("model service provider is required")
	}
	if config.ModelSettings.ServiceAPIKey == "" {
		return fmt.Errorf("model service API key is required")
	}

	// Validate logging settings
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[config.LoggingSettings.LogLevel] {
		return fmt.Errorf("invalid log level")
	}

	// Validate operational mode
	validModes := map[string]bool{
		"suggest": true,
		"deploy":  true,
	}
	if !validModes[config.OperationalMode] {
		return fmt.Errorf("invalid operational mode")
	}

	// Validate repository settings
	if config.RepositorySettings.RepositoryPath == "" {
		return fmt.Errorf("repository path is required")
	}

	return nil
}
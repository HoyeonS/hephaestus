package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"gopkg.in/yaml.v3"
)

// loadConfigurationFromFile loads configuration from a YAML configuration file
func (cm *ConfigurationManager) loadConfigurationFromFile(configPath string) error {
	if configPath == "" {
		return nil // Skip if no config file specified
	}

	fileContents, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	configurationObject := &hephaestus.SystemConfiguration{}
	if err := yaml.Unmarshal(fileContents, configurationObject); err != nil {
		return fmt.Errorf("failed to parse configuration file: %w", err)
	}

	cm.currentConfig = configurationObject
	return nil
}

// loadConfigurationFromEnvironment loads configuration from environment variables
func (cm *ConfigurationManager) loadConfigurationFromEnvironment() error {
	if cm.currentConfig == nil {
		cm.currentConfig = &hephaestus.SystemConfiguration{}
	}

	// Load Remote Repository configuration from environment
	cm.loadRemoteRepositoryConfigurationFromEnvironment(&cm.currentConfig.RemoteSettings)

	// Load Model configuration from environment
	cm.loadModelConfigurationFromEnvironment(&cm.currentConfig.ModelSettings)

	// Load Log configuration from environment
	cm.loadLoggingConfigurationFromEnvironment(&cm.currentConfig.LoggingSettings)

	// Load Repository configuration from environment
	cm.loadRepositoryConfigurationFromEnvironment(&cm.currentConfig.RepositorySettings)

	// Load Operation Mode from environment
	if mode := os.Getenv("HEPHAESTUS_MODE"); mode != "" {
		cm.currentConfig.OperationalMode = mode
	}

	return nil
}

// loadRemoteRepositoryConfigurationFromEnvironment loads remote repository configuration from environment variables
func (cm *ConfigurationManager) loadRemoteRepositoryConfigurationFromEnvironment(config *hephaestus.RemoteRepositoryConfiguration) {
	if token := os.Getenv("REMOTE_TOKEN"); token != "" {
		config.AuthToken = token
	}
	if owner := os.Getenv("REMOTE_OWNER"); owner != "" {
		config.RepositoryOwner = owner
	}
	if repo := os.Getenv("REMOTE_REPO"); repo != "" {
		config.RepositoryName = repo
	}
	if branch := os.Getenv("REMOTE_BRANCH"); branch != "" {
		config.TargetBranch = branch
	}
}

// loadModelConfigurationFromEnvironment loads model configuration from environment variables
func (cm *ConfigurationManager) loadModelConfigurationFromEnvironment(config *hephaestus.ModelServiceConfiguration) {
	if provider := os.Getenv("MODEL_PROVIDER"); provider != "" {
		config.ServiceProvider = provider
	}
	if apiKey := os.Getenv("MODEL_API_KEY"); apiKey != "" {
		config.ServiceAPIKey = apiKey
	}
	if version := os.Getenv("MODEL_VERSION"); version != "" {
		config.ModelVersion = version
	}
}

// loadLoggingConfigurationFromEnvironment loads logging configuration from environment variables
func (cm *ConfigurationManager) loadLoggingConfigurationFromEnvironment(config *hephaestus.LoggingConfiguration) {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.LogLevel = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.OutputFormat = format
	}
}

// loadRepositoryConfigurationFromEnvironment loads repository configuration from environment variables
func (cm *ConfigurationManager) loadRepositoryConfigurationFromEnvironment(config *hephaestus.RepositoryConfiguration) {
	if path := os.Getenv("REPO_PATH"); path != "" {
		config.RepositoryPath = path
	}
	if maxFiles := os.Getenv("REPO_FILE_LIMIT"); maxFiles != "" {
		if val, err := strconv.Atoi(maxFiles); err == nil {
			config.FileLimit = val
		}
	}
	if maxFileSize := os.Getenv("REPO_FILE_SIZE_LIMIT"); maxFileSize != "" {
		if val, err := strconv.ParseInt(maxFileSize, 10, 64); err == nil {
			config.FileSizeLimit = val
		}
	}
}

// SaveConfiguration persists the current configuration to a file
func (cm *ConfigurationManager) SaveConfiguration(configPath string) error {
	if cm.currentConfig == nil {
		return fmt.Errorf("no configuration available to save")
	}

	// Ensure configuration directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %w", err)
	}

	// Marshal configuration to YAML format
	configData, err := yaml.Marshal(cm.currentConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write configuration to file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// GetDefaultConfigurationFilePath returns the system's default configuration file path
func GetDefaultConfigurationFilePath() string {
	if customPath := os.Getenv("HEPHAESTUS_CONFIG_PATH"); customPath != "" {
		return customPath
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "hephaestus", "config.yaml")
}

// GetDefaultConfiguration returns the default system configuration
func GetDefaultConfiguration() *hephaestus.SystemConfiguration {
	return &hephaestus.SystemConfiguration{
		RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
			TargetBranch: "main",
		},
		LoggingSettings: hephaestus.LoggingConfiguration{
			LogLevel:     "info",
			OutputFormat: "json",
		},
		RepositorySettings: hephaestus.RepositoryConfiguration{
			FileLimit:     10000,
			FileSizeLimit: 1 << 20, // 1MB
		},
		OperationalMode: "suggest",
	}
} 
} 
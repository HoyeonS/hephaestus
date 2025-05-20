package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"gopkg.in/yaml.v3"
)

// ConfigurationManager handles system configuration
type ConfigurationManager struct {
	config *hephaestus.SystemConfiguration
	path   string
}

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager(configPath string) *ConfigurationManager {
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}
	return &ConfigurationManager{
		path: configPath,
	}
}

// Get returns the current configuration
func (m *ConfigurationManager) Get() *hephaestus.SystemConfiguration {
	return m.config
}

// Set updates the configuration
func (m *ConfigurationManager) Set(config *hephaestus.SystemConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if err := hephaestus.ValidateSystemConfiguration(config); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	m.config = config
	return nil
}

// LoadConfiguration loads configuration from file and environment
func (m *ConfigurationManager) LoadConfiguration() error {
	// Initialize with default configuration
	if m.config == nil {
		m.config = GetDefaultConfiguration()
	}

	// Load from file if it exists
	if err := m.loadConfigurationFromFile(); err != nil {
		return err
	}

	// Load from environment variables
	if err := m.loadConfigurationFromEnvironment(); err != nil {
		return err
	}

	return nil
}

// loadConfigurationFromFile loads configuration from a YAML file
func (m *ConfigurationManager) loadConfigurationFromFile() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create with default configuration
			return m.createDefaultConfiguration()
		}
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	var config hephaestus.SystemConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse configuration file: %v", err)
	}

	return m.Set(&config)
}

// loadConfigurationFromEnvironment loads configuration from environment variables
func (m *ConfigurationManager) loadConfigurationFromEnvironment() error {
	// Load Remote Repository configuration
	m.loadRemoteRepositoryConfigurationFromEnvironment(&m.config.RemoteSettings)

	// Load Model configuration
	m.loadModelConfigurationFromEnvironment(&m.config.ModelSettings)

	// Load Log configuration
	m.loadLoggingConfigurationFromEnvironment(&m.config.LoggingSettings)

	// Load Repository configuration
	m.loadRepositoryConfigurationFromEnvironment(&m.config.RepositorySettings)

	// Load Operation Mode
	if mode := os.Getenv("HEPHAESTUS_MODE"); mode != "" {
		m.config.OperationalMode = mode
	}

	return nil
}

// loadRemoteRepositoryConfigurationFromEnvironment loads remote repository configuration
func (m *ConfigurationManager) loadRemoteRepositoryConfigurationFromEnvironment(config *hephaestus.RemoteRepositoryConfiguration) {
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

// loadModelConfigurationFromEnvironment loads model configuration
func (m *ConfigurationManager) loadModelConfigurationFromEnvironment(config *hephaestus.ModelServiceConfiguration) {
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

// loadLoggingConfigurationFromEnvironment loads logging configuration
func (m *ConfigurationManager) loadLoggingConfigurationFromEnvironment(config *hephaestus.LoggingConfiguration) {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.LogLevel = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.OutputFormat = format
	}
}

// loadRepositoryConfigurationFromEnvironment loads repository configuration
func (m *ConfigurationManager) loadRepositoryConfigurationFromEnvironment(config *hephaestus.RepositoryConfiguration) {
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

// SaveConfiguration saves the current configuration to a YAML file
func (m *ConfigurationManager) SaveConfiguration() error {
	if m.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	// Ensure directory exists
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %v", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	if err := os.WriteFile(m.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write configuration file: %v", err)
	}

	return nil
}

// createDefaultConfiguration creates a new configuration file with default values
func (m *ConfigurationManager) createDefaultConfiguration() error {
	defaultConfig := GetDefaultConfiguration()
	if err := m.Set(defaultConfig); err != nil {
		return err
	}
	return m.SaveConfiguration()
}

// GetDefaultConfigPath returns the default path for the configuration file
func GetDefaultConfigPath() string {
	if customPath := os.Getenv("HEPHAESTUS_CONFIG_PATH"); customPath != "" {
		return customPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".hephaestus/config.yaml"
	}
	return filepath.Join(home, ".hephaestus", "config.yaml")
}

// GetDefaultConfiguration returns a default configuration
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
			FileSizeLimit: 1048576, // 1MB
		},
		OperationalMode: "suggest",
	}
}

// ValidateConfiguration validates the configuration file exists and is readable
func (m *ConfigurationManager) ValidateConfiguration() error {
	info, err := os.Stat(m.path)
	if err != nil {
		return fmt.Errorf("configuration file not found: %v", err)
	}

	// Check file permissions
	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("configuration file has incorrect permissions: %v", info.Mode().Perm())
	}

	return nil
}
 
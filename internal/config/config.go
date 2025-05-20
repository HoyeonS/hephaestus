package config

import (
	"fmt"
	"os"
	"path/filepath"

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

// LoadConfiguration loads configuration from a YAML file
func (m *ConfigurationManager) LoadConfiguration() error {
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
 
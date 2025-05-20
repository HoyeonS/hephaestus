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
}

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{}
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
func (m *ConfigurationManager) LoadConfiguration(path string) error {
	if path == "" {
		path = GetDefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	var config hephaestus.SystemConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse configuration file: %v", err)
	}

	return m.Set(&config)
}

// SaveConfiguration saves the current configuration to a YAML file
func (m *ConfigurationManager) SaveConfiguration(path string) error {
	if m.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	if path == "" {
		path = GetDefaultConfigPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %v", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %v", err)
	}

	return nil
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

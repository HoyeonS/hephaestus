package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"gopkg.in/yaml.v3"
)

// ConfigurationManager handles system configuration
type ConfigurationManager struct {
	nativeConfig *hephaestus.SystemConfiguration
	clientConfig *hephaestus.ClientConfiguration
	nativePath   string
	clientPath   string
}

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager(nativePath, clientPath string) *ConfigurationManager {
	if nativePath == "" {
		nativePath = GetDefaultNativeConfigPath()
	}
	if clientPath == "" {
		clientPath = GetDefaultClientConfigPath()
	}
	return &ConfigurationManager{
		nativePath: nativePath,
		clientPath: clientPath,
	}
}

// GetNativeConfig returns the current native configuration
func (m *ConfigurationManager) GetNativeConfig() *hephaestus.SystemConfiguration {
	return m.nativeConfig
}

// GetClientConfig returns the current client configuration
func (m *ConfigurationManager) GetClientConfig() *hephaestus.ClientConfiguration {
	return m.clientConfig
}

// SetNativeConfig updates the native configuration
func (m *ConfigurationManager) SetNativeConfig(config *hephaestus.SystemConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if err := hephaestus.ValidateSystemConfiguration(config); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	m.nativeConfig = config
	return nil
}

// SetClientConfig updates the client configuration
func (m *ConfigurationManager) SetClientConfig(config *hephaestus.ClientConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if err := hephaestus.ValidateClientConfiguration(config); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	m.clientConfig = config
	return nil
}

// LoadConfiguration loads both native and client configurations from files
func (m *ConfigurationManager) LoadConfiguration() error {
	// Initialize with default configurations
	if m.nativeConfig == nil {
		m.nativeConfig = GetDefaultNativeConfiguration()
	}
	if m.clientConfig == nil {
		m.clientConfig = GetDefaultClientConfiguration()
	}

	// Load native configuration from file
	if err := m.loadNativeConfigurationFromFile(); err != nil {
		return fmt.Errorf("failed to load native configuration: %v", err)
	}

	// Load client configuration from file
	if err := m.loadClientConfigurationFromFile(); err != nil {
		return fmt.Errorf("failed to load client configuration: %v", err)
	}

	return nil
}

// loadNativeConfigurationFromFile loads native configuration from a YAML file
func (m *ConfigurationManager) loadNativeConfigurationFromFile() error {
	data, err := os.ReadFile(m.nativePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create with default configuration
			return m.createDefaultNativeConfiguration()
		}
		return fmt.Errorf("failed to read native configuration file: %v", err)
	}

	var config hephaestus.SystemConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse native configuration file: %v", err)
	}

	return m.SetNativeConfig(&config)
}

// loadClientConfigurationFromFile loads client configuration from a YAML file
func (m *ConfigurationManager) loadClientConfigurationFromFile() error {
	data, err := os.ReadFile(m.clientPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create with default configuration
			return m.createDefaultClientConfiguration()
		}
		return fmt.Errorf("failed to read client configuration file: %v", err)
	}

	var config hephaestus.ClientConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse client configuration file: %v", err)
	}

	return m.SetClientConfig(&config)
}

// SaveConfiguration saves both native and client configurations to YAML files
func (m *ConfigurationManager) SaveConfiguration() error {
	if m.nativeConfig == nil || m.clientConfig == nil {
		return fmt.Errorf("no configuration to save")
	}

	// Save native configuration
	if err := m.saveNativeConfiguration(); err != nil {
		return fmt.Errorf("failed to save native configuration: %v", err)
	}

	// Save client configuration
	if err := m.saveClientConfiguration(); err != nil {
		return fmt.Errorf("failed to save client configuration: %v", err)
	}

	return nil
}

// saveNativeConfiguration saves the native configuration to a YAML file
func (m *ConfigurationManager) saveNativeConfiguration() error {
	// Ensure directory exists
	dir := filepath.Dir(m.nativePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create native configuration directory: %v", err)
	}

	data, err := yaml.Marshal(m.nativeConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal native configuration: %v", err)
	}

	if err := os.WriteFile(m.nativePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write native configuration file: %v", err)
	}

	return nil
}

// saveClientConfiguration saves the client configuration to a YAML file
func (m *ConfigurationManager) saveClientConfiguration() error {
	// Ensure directory exists
	dir := filepath.Dir(m.clientPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create client configuration directory: %v", err)
	}

	data, err := yaml.Marshal(m.clientConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal client configuration: %v", err)
	}

	if err := os.WriteFile(m.clientPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write client configuration file: %v", err)
	}

	return nil
}

// createDefaultNativeConfiguration creates a new native configuration file with default values
func (m *ConfigurationManager) createDefaultNativeConfiguration() error {
	defaultConfig := GetDefaultNativeConfiguration()
	if err := m.SetNativeConfig(defaultConfig); err != nil {
		return err
	}
	return m.saveNativeConfiguration()
}

// createDefaultClientConfiguration creates a new client configuration file with default values
func (m *ConfigurationManager) createDefaultClientConfiguration() error {
	defaultConfig := GetDefaultClientConfiguration()
	if err := m.SetClientConfig(defaultConfig); err != nil {
		return err
	}
	return m.saveClientConfiguration()
}

// GetDefaultNativeConfigPath returns the default path for the native configuration file
func GetDefaultNativeConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".hephaestus/native_config.yaml"
	}
	return filepath.Join(home, ".hephaestus", "native_config.yaml")
}

// GetDefaultClientConfigPath returns the default path for the client configuration file
func GetDefaultClientConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".hephaestus/client_config.yaml"
	}
	return filepath.Join(home, ".hephaestus", "client_config.yaml")
}

// GetDefaultNativeConfiguration returns a default native configuration
func GetDefaultNativeConfiguration() *hephaestus.SystemConfiguration {
	return &hephaestus.SystemConfiguration{
		Model: hephaestus.ModelConfiguration{
			Provider:    "openai",
			ModelVersion: "gpt-4",
		},
		LogSettings: hephaestus.LogProcessingConfiguration{
			ThresholdLevel:  "error",
			ThresholdCount:  5,
			ThresholdWindow: 300 * time.Second,
		},
		OperationalMode: "suggest",
	}
}

// GetDefaultClientConfiguration returns a default client configuration
func GetDefaultClientConfiguration() *hephaestus.ClientConfiguration {
	return &hephaestus.ClientConfiguration{
		Logging: hephaestus.LoggingConfiguration{
			Level:  "info",
			Format: "json",
		},
		Repository: hephaestus.RepositoryConfiguration{
			FileLimit:     10000,
			FileSizeLimit: 1048576, // 1MB
		},
		BaseURL: "http://localhost:8080",
	}
}

// ValidateConfiguration validates both configuration files exist and are readable
func (m *ConfigurationManager) ValidateConfiguration() error {
	// Validate native configuration
	nativeInfo, err := os.Stat(m.nativePath)
	if err != nil {
		return fmt.Errorf("native configuration file not found: %v", err)
	}
	if nativeInfo.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("native configuration file has incorrect permissions: %v", nativeInfo.Mode().Perm())
	}

	// Validate client configuration
	clientInfo, err := os.Stat(m.clientPath)
	if err != nil {
		return fmt.Errorf("client configuration file not found: %v", err)
	}
	if clientInfo.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("client configuration file has incorrect permissions: %v", clientInfo.Mode().Perm())
	}

	return nil
}
 
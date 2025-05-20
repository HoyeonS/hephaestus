package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationManager(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "hephaestus-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")

	t.Run("New", func(t *testing.T) {
		manager := NewConfigurationManager(configPath)
		assert.NotNil(t, manager)
		assert.Equal(t, configPath, manager.path)
	})

	t.Run("Default Path", func(t *testing.T) {
		manager := NewConfigurationManager("")
		assert.NotEmpty(t, manager.path)
	})

	t.Run("Set and Get", func(t *testing.T) {
		manager := NewConfigurationManager(configPath)
		config := &hephaestus.SystemConfiguration{
			RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
				AuthToken:       "token",
				RepositoryOwner: "owner",
				RepositoryName:  "repo",
				TargetBranch:    "main",
			},
			ModelSettings: hephaestus.ModelServiceConfiguration{
				ServiceProvider: "openai",
				ServiceAPIKey:   "key",
				ModelVersion:    "gpt-4",
			},
			LoggingSettings: hephaestus.LoggingConfiguration{
				LogLevel:     "info",
				OutputFormat: "json",
			},
			RepositorySettings: hephaestus.RepositoryConfiguration{
				RepositoryPath: "/path/to/repo",
				FileLimit:      10000,
				FileSizeLimit:  1048576,
			},
			OperationalMode: "suggest",
		}
		err := manager.Set(config)
		assert.NoError(t, err)

		got := manager.Get()
		assert.Equal(t, config, got)
	})

	t.Run("Load and Save", func(t *testing.T) {
		manager := NewConfigurationManager(configPath)
		config := &hephaestus.SystemConfiguration{
			RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
				AuthToken:       "token",
				RepositoryOwner: "owner",
				RepositoryName:  "repo",
				TargetBranch:    "main",
			},
			ModelSettings: hephaestus.ModelServiceConfiguration{
				ServiceProvider: "openai",
				ServiceAPIKey:   "key",
				ModelVersion:    "gpt-4",
			},
			LoggingSettings: hephaestus.LoggingConfiguration{
				LogLevel:     "info",
				OutputFormat: "json",
			},
			RepositorySettings: hephaestus.RepositoryConfiguration{
				RepositoryPath: "/path/to/repo",
				FileLimit:      10000,
				FileSizeLimit:  1048576,
			},
			OperationalMode: "suggest",
		}

		// Save configuration
		err := manager.Set(config)
		require.NoError(t, err)
		err = manager.SaveConfiguration()
		require.NoError(t, err)

		// Create new manager and load configuration
		manager2 := NewConfigurationManager(configPath)
		err = manager2.LoadConfiguration()
		require.NoError(t, err)

		// Compare configurations
		assert.Equal(t, config, manager2.Get())
	})

	t.Run("Create Default Configuration", func(t *testing.T) {
		manager := NewConfigurationManager(configPath)
		err := manager.LoadConfiguration()
		require.NoError(t, err)

		config := manager.Get()
		assert.NotNil(t, config)
		assert.Equal(t, "main", config.RemoteSettings.TargetBranch)
		assert.Equal(t, "info", config.LoggingSettings.LogLevel)
		assert.Equal(t, "json", config.LoggingSettings.OutputFormat)
		assert.Equal(t, 10000, config.RepositorySettings.FileLimit)
		assert.Equal(t, int64(1048576), config.RepositorySettings.FileSizeLimit)
		assert.Equal(t, "suggest", config.OperationalMode)
	})

	t.Run("Validate Configuration", func(t *testing.T) {
		manager := NewConfigurationManager(configPath)
		err := manager.LoadConfiguration()
		require.NoError(t, err)

		err = manager.ValidateConfiguration()
		assert.NoError(t, err)

		// Test with incorrect permissions
		err = os.Chmod(configPath, 0777)
		require.NoError(t, err)
		err = manager.ValidateConfiguration()
		assert.Error(t, err)
	})
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *hephaestus.SystemConfiguration
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil config",
			config:      nil,
			wantErr:     true,
			errContains: "configuration is nil",
		},
		{
			name: "missing GitHub token",
			config: &hephaestus.SystemConfiguration{
				RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
					AuthToken:       "",
					RepositoryOwner: "owner",
					RepositoryName:  "repo",
					TargetBranch:    "main",
				},
				ModelSettings: hephaestus.ModelServiceConfiguration{
					ServiceProvider: "openai",
					ServiceAPIKey:   "key",
					ModelVersion:    "gpt-4",
				},
				LoggingSettings: hephaestus.LoggingConfiguration{
					LogLevel:     "info",
					OutputFormat: "json",
				},
				RepositorySettings: hephaestus.RepositoryConfiguration{
					RepositoryPath: "/path/to/repo",
				},
				OperationalMode: "suggest",
			},
			wantErr:     true,
			errContains: "remote auth token is required",
		},
		{
			name: "missing AI provider",
			config: &hephaestus.SystemConfiguration{
				RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
					AuthToken:       "token",
					RepositoryOwner: "owner",
					RepositoryName:  "repo",
					TargetBranch:    "main",
				},
				ModelSettings: hephaestus.ModelServiceConfiguration{
					ServiceProvider: "",
					ServiceAPIKey:   "key",
					ModelVersion:    "gpt-4",
				},
				LoggingSettings: hephaestus.LoggingConfiguration{
					LogLevel:     "info",
					OutputFormat: "json",
				},
				RepositorySettings: hephaestus.RepositoryConfiguration{
					RepositoryPath: "/path/to/repo",
				},
				OperationalMode: "suggest",
			},
			wantErr:     true,
			errContains: "model service provider is required",
		},
		{
			name: "invalid log level",
			config: &hephaestus.SystemConfiguration{
				RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
					AuthToken:       "token",
					RepositoryOwner: "owner",
					RepositoryName:  "repo",
					TargetBranch:    "main",
				},
				ModelSettings: hephaestus.ModelServiceConfiguration{
					ServiceProvider: "openai",
					ServiceAPIKey:   "key",
					ModelVersion:    "gpt-4",
				},
				LoggingSettings: hephaestus.LoggingConfiguration{
					LogLevel:     "invalid",
					OutputFormat: "json",
				},
				RepositorySettings: hephaestus.RepositoryConfiguration{
					RepositoryPath: "/path/to/repo",
				},
				OperationalMode: "suggest",
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "valid config",
			config: &hephaestus.SystemConfiguration{
				RemoteSettings: hephaestus.RemoteRepositoryConfiguration{
					AuthToken:       "token",
					RepositoryOwner: "owner",
					RepositoryName:  "repo",
					TargetBranch:    "main",
				},
				ModelSettings: hephaestus.ModelServiceConfiguration{
					ServiceProvider: "openai",
					ServiceAPIKey:   "key",
					ModelVersion:    "gpt-4",
				},
				LoggingSettings: hephaestus.LoggingConfiguration{
					LogLevel:     "info",
					OutputFormat: "json",
				},
				RepositorySettings: hephaestus.RepositoryConfiguration{
					RepositoryPath: "/path/to/repo",
				},
				OperationalMode: "suggest",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewConfigurationManager("")
			err := manager.Set(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// func TestConfigFile(t *testing.T) {
// 	// Create temporary directory for test files
// 	tmpDir, err := os.MkdirTemp("", "hephaestus-test-*")
// 	require.NoError(t, err)
// 	defer os.RemoveAll(tmpDir)

// 	configPath := filepath.Join(tmpDir, "config.yaml")

// 	// Create test configuration
// 	config := &hephaestus.Config{
// 		Repository: &hephaestus.RepositoryConfig{
// 			Token:    "token",
// 			URL:      "URL",
// 			Type:     "type",
// 			Branch:   "main",
// 			BasePath: "/",
// 		},
// 		Model: &hephaestus.ModelConfig{
// 			Provider: "aiprovider",
// 			APIKey:   "key",
// 			Model:    "model",
// 		},
// 		Log: &hephaestus.LogConfig{
// 			Level:  "info",
// 			Output: "json",
// 		},
// 	}

// 	t.Run("Save and Load", func(t *testing.T) {
// 		manager := NewConfigurationManager()
// 		systemConfig, err := SystemConfigurationFactory(config, "")
// 		require.NoError(t, err)

// 		err = manager.Set(systemConfig)
// 		require.NoError(t, err)

// 		// Save configuration
// 		err = manager.SaveConfiguration(configPath)
// 		require.NoError(t, err)

// 		// Load configuration
// 		manager2 := NewConfigurationManager()
// 		err = manager2.loadConfigurationFromFile(configPath)
// 		require.NoError(t, err)

// 		// Compare configurations
// 		assert.Equal(t, config, manager2.Get())
// 	})
// }

// func TestEnvironmentVariables(t *testing.T) {
// 	// Set environment variables
// 	envVars := map[string]string{
// 		"GITHUB_TOKEN":       "env-token",
// 		"GITHUB_OWNER":       "env-owner",
// 		"GITHUB_REPO":        "env-repo",
// 		"GITHUB_BRANCH":      "env-branch",
// 		"AI_PROVIDER":        "env-provider",
// 		"AI_API_KEY":         "env-key",
// 		"LOG_LEVEL":          "info",
// 		"LOG_FORMAT":         "json",
// 		"REPO_PATH":          "/env/path",
// 		"REPO_MAX_FILES":     "5000",
// 		"REPO_MAX_FILE_SIZE": "2097152",
// 		"HEPHAESTUS_MODE":    "deploy",
// 	}

// 	for k, v := range envVars {
// 		os.Setenv(k, v)
// 		defer os.Unsetenv(k)
// 	}

// 	t.Run("Load from Environment", func(t *testing.T) {
// 		manager := New()
// 		err := manager.loadFromEnv()
// 		require.NoError(t, err)

// 		config := manager.Get()
// 		assert.Equal(t, "env-token", config.GitHub.Token)
// 		assert.Equal(t, "env-owner", config.GitHub.Owner)
// 		assert.Equal(t, "env-repo", config.GitHub.Repository)
// 		assert.Equal(t, "env-branch", config.GitHub.Branch)
// 		assert.Equal(t, "env-provider", config.AI.Provider)
// 		assert.Equal(t, "env-key", config.AI.APIKey)
// 		assert.Equal(t, "info", config.Log.Level)
// 		assert.Equal(t, "json", config.Log.Format)
// 		assert.Equal(t, "/env/path", config.Repository.Path)
// 		assert.Equal(t, 5000, config.Repository.MaxFiles)
// 		assert.Equal(t, int64(2097152), config.Repository.MaxFileSize)
// 		assert.Equal(t, "deploy", config.Mode)
// 	})
// }

// func TestDefaultConfig(t *testing.T) {
// 	config := GetDefaultConfig()
// 	assert.NotNil(t, config)
// 	assert.Equal(t, "main", config.GitHub.Branch)
// 	assert.Equal(t, "info", config.Log.Level)
// 	assert.Equal(t, "json", config.Log.Format)
// 	assert.Equal(t, 10000, config.Repository.MaxFiles)
// 	assert.Equal(t, int64(1<<20), config.Repository.MaxFileSize)
// 	assert.Equal(t, "suggest", config.Mode)
// }

// func TestConfigFilePath(t *testing.T) {
// 	t.Run("Default Path", func(t *testing.T) {
// 		home := os.Getenv("HOME")
// 		expected := filepath.Join(home, ".config", "hephaestus", "config.yaml")
// 		assert.Equal(t, expected, GetConfigFilePath())
// 	})

// 	t.Run("Custom Path", func(t *testing.T) {
// 		customPath := "/custom/path/config.yaml"
// 		os.Setenv("HEPHAESTUS_CONFIG_PATH", customPath)
// 		defer os.Unsetenv("HEPHAESTUS_CONFIG_PATH")

// 		assert.Equal(t, customPath, GetConfigFilePath())
// 	})
// }

// func TestLoadConfig(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		path     string
// 		expected *hephaestus.Config
// 		wantErr  bool
// 	}{
// 		{
// 			name: "valid config",
// 			path: "testdata/valid_config.yaml",
// 			expected: &hephaestus.Config{
// 				LogLevel:  "info",
// 				LogOutput: "stdout",
// 				NodeID:    "test-node",
// 				Repository: &hephaestus.RepositoryConfig{
// 					Owner:    "HoyeonS",
// 					Name:     "hephaestus",
// 					Token:    "test-token",
// 					BasePath: "/base/path",
// 					Branch:   "main",
// 				},
// 				Model: &hephaestus.ModelConfig{
// 					Version: "v1",
// 					APIKey:  "test-key",
// 					BaseURL: "https://api.example.com",
// 					Timeout: 30,
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:     "invalid path",
// 			path:     "testdata/nonexistent.yaml",
// 			expected: nil,
// 			wantErr:  true,
// 		},
// 		{
// 			name:     "invalid yaml",
// 			path:     "testdata/invalid_config.yaml",
// 			expected: nil,
// 			wantErr:  true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := LoadConfig(tt.path)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				assert.Nil(t, got)
// 				return
// 			}

// 			assert.NoError(t, err)
// 			assert.NotNil(t, got)
// 			assert.Equal(t, tt.expected, got)
// 		})
// 	}
// }

// func TestValidateConfig(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		config  *hephaestus.Config
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid config",
// 			config: &hephaestus.Config{
// 				LogLevel:  "info",
// 				LogOutput: "stdout",
// 				NodeID:    "test-node",
// 				Repository: &hephaestus.RepositoryConfig{
// 					Owner:    "HoyeonS",
// 					Name:     "hephaestus",
// 					Token:    "test-token",
// 					BasePath: "/base/path",
// 					Branch:   "main",
// 				},
// 				Model: &hephaestus.ModelConfig{
// 					Version: "v1",
// 					APIKey:  "test-key",
// 					BaseURL: "https://api.example.com",
// 					Timeout: 30,
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "nil config",
// 			config:  nil,
// 			wantErr: true,
// 		},
// 		{
// 			name: "missing log level",
// 			config: &hephaestus.Config{
// 				LogOutput: "stdout",
// 				NodeID:    "test-node",
// 				Repository: &hephaestus.RepositoryConfig{
// 					Owner:    "HoyeonS",
// 					Name:     "hephaestus",
// 					Token:    "test-token",
// 					BasePath: "/base/path",
// 					Branch:   "main",
// 				},
// 				Model: &hephaestus.ModelConfig{
// 					Version: "v1",
// 					APIKey:  "test-key",
// 					BaseURL: "https://api.example.com",
// 					Timeout: 30,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "missing repository config",
// 			config: &hephaestus.Config{
// 				LogLevel:  "info",
// 				LogOutput: "stdout",
// 				NodeID:    "test-node",
// 				Model: &hephaestus.ModelConfig{
// 					Version: "v1",
// 					APIKey:  "test-key",
// 					BaseURL: "https://api.example.com",
// 					Timeout: 30,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "missing model config",
// 			config: &hephaestus.Config{
// 				LogLevel:  "info",
// 				LogOutput: "stdout",
// 				NodeID:    "test-node",
// 				Repository: &hephaestus.RepositoryConfig{
// 					Owner:    "HoyeonS",
// 					Name:     "hephaestus",
// 					Token:    "test-token",
// 					BasePath: "/base/path",
// 					Branch:   "main",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := ValidateConfig(tt.config)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				return
// 			}
// 			assert.NoError(t, err)
// 		})
// 	}
// }

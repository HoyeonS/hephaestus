package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		manager := New()
		assert.NotNil(t, manager)
	})

	t.Run("Set and Get", func(t *testing.T) {
		manager := New()
		config := &hephaestus.Config{
			GitHub: hephaestus.GitHubConfig{
				Token:      "token",
				Owner:      "owner",
				Repository: "repo",
				Branch:     "main",
			},
			AI: hephaestus.AIConfig{
				Provider: "openai",
				APIKey:   "key",
			},
			Log: hephaestus.LogConfig{
				Level:  "info",
				Format: "json",
			},
			Repository: hephaestus.RepositoryConfig{
				Path:        "/path/to/repo",
				MaxFiles:    10000,
				MaxFileSize: 1 << 20,
			},
			Mode: "suggest",
		}

		err := manager.Set(config)
		assert.NoError(t, err)

		got := manager.Get()
		assert.Equal(t, config, got)
	})
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *hephaestus.Config
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
			config: &hephaestus.Config{
				GitHub: hephaestus.GitHubConfig{
					Owner:      "owner",
					Repository: "repo",
				},
				AI: hephaestus.AIConfig{
					Provider: "openai",
					APIKey:   "key",
				},
				Log: hephaestus.LogConfig{
					Level:  "info",
					Format: "json",
				},
				Repository: hephaestus.RepositoryConfig{
					Path: "/path",
				},
				Mode: "suggest",
			},
			wantErr:     true,
			errContains: "GitHub token is required",
		},
		{
			name: "missing AI provider",
			config: &hephaestus.Config{
				GitHub: hephaestus.GitHubConfig{
					Token:      "token",
					Owner:      "owner",
					Repository: "repo",
				},
				AI: hephaestus.AIConfig{
					APIKey: "key",
				},
				Log: hephaestus.LogConfig{
					Level:  "info",
					Format: "json",
				},
				Repository: hephaestus.RepositoryConfig{
					Path: "/path",
				},
				Mode: "suggest",
			},
			wantErr:     true,
			errContains: "AI provider is required",
		},
		{
			name: "invalid log level",
			config: &hephaestus.Config{
				GitHub: hephaestus.GitHubConfig{
					Token:      "token",
					Owner:      "owner",
					Repository: "repo",
				},
				AI: hephaestus.AIConfig{
					Provider: "openai",
					APIKey:   "key",
				},
				Log: hephaestus.LogConfig{
					Level:  "invalid",
					Format: "json",
				},
				Repository: hephaestus.RepositoryConfig{
					Path: "/path",
				},
				Mode: "suggest",
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid mode",
			config: &hephaestus.Config{
				GitHub: hephaestus.GitHubConfig{
					Token:      "token",
					Owner:      "owner",
					Repository: "repo",
				},
				AI: hephaestus.AIConfig{
					Provider: "openai",
					APIKey:   "key",
				},
				Log: hephaestus.LogConfig{
					Level:  "info",
					Format: "json",
				},
				Repository: hephaestus.RepositoryConfig{
					Path: "/path",
				},
				Mode: "invalid",
			},
			wantErr:     true,
			errContains: "invalid mode",
		},
		{
			name: "valid config",
			config: &hephaestus.Config{
				GitHub: hephaestus.GitHubConfig{
					Token:      "token",
					Owner:      "owner",
					Repository: "repo",
					Branch:     "main",
				},
				AI: hephaestus.AIConfig{
					Provider: "openai",
					APIKey:   "key",
				},
				Log: hephaestus.LogConfig{
					Level:  "info",
					Format: "json",
				},
				Repository: hephaestus.RepositoryConfig{
					Path:        "/path",
					MaxFiles:    10000,
					MaxFileSize: 1 << 20,
				},
				Mode: "suggest",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := New()
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

func TestConfigFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "hephaestus-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create test configuration
	config := &hephaestus.Config{
		GitHub: hephaestus.GitHubConfig{
			Token:      "token",
			Owner:      "owner",
			Repository: "repo",
			Branch:     "main",
		},
		AI: hephaestus.AIConfig{
			Provider: "openai",
			APIKey:   "key",
		},
		Log: hephaestus.LogConfig{
			Level:  "info",
			Format: "json",
		},
		Repository: hephaestus.RepositoryConfig{
			Path:        "/path/to/repo",
			MaxFiles:    10000,
			MaxFileSize: 1 << 20,
		},
		Mode: "suggest",
	}

	t.Run("Save and Load", func(t *testing.T) {
		manager := New()
		err := manager.Set(config)
		require.NoError(t, err)

		// Save configuration
		err = manager.Save(configPath)
		require.NoError(t, err)

		// Load configuration
		manager2 := New()
		err = manager2.Load(configPath)
		require.NoError(t, err)

		// Compare configurations
		assert.Equal(t, config, manager2.Get())
	})
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"GITHUB_TOKEN":        "env-token",
		"GITHUB_OWNER":        "env-owner",
		"GITHUB_REPO":         "env-repo",
		"GITHUB_BRANCH":       "env-branch",
		"AI_PROVIDER":         "env-provider",
		"AI_API_KEY":         "env-key",
		"LOG_LEVEL":          "info",
		"LOG_FORMAT":         "json",
		"REPO_PATH":          "/env/path",
		"REPO_MAX_FILES":     "5000",
		"REPO_MAX_FILE_SIZE": "2097152",
		"HEPHAESTUS_MODE":    "deploy",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	t.Run("Load from Environment", func(t *testing.T) {
		manager := New()
		err := manager.loadFromEnv()
		require.NoError(t, err)

		config := manager.Get()
		assert.Equal(t, "env-token", config.GitHub.Token)
		assert.Equal(t, "env-owner", config.GitHub.Owner)
		assert.Equal(t, "env-repo", config.GitHub.Repository)
		assert.Equal(t, "env-branch", config.GitHub.Branch)
		assert.Equal(t, "env-provider", config.AI.Provider)
		assert.Equal(t, "env-key", config.AI.APIKey)
		assert.Equal(t, "info", config.Log.Level)
		assert.Equal(t, "json", config.Log.Format)
		assert.Equal(t, "/env/path", config.Repository.Path)
		assert.Equal(t, 5000, config.Repository.MaxFiles)
		assert.Equal(t, int64(2097152), config.Repository.MaxFileSize)
		assert.Equal(t, "deploy", config.Mode)
	})
}

func TestDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "main", config.GitHub.Branch)
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "json", config.Log.Format)
	assert.Equal(t, 10000, config.Repository.MaxFiles)
	assert.Equal(t, int64(1<<20), config.Repository.MaxFileSize)
	assert.Equal(t, "suggest", config.Mode)
}

func TestConfigFilePath(t *testing.T) {
	t.Run("Default Path", func(t *testing.T) {
		home := os.Getenv("HOME")
		expected := filepath.Join(home, ".config", "hephaestus", "config.yaml")
		assert.Equal(t, expected, GetConfigFilePath())
	})

	t.Run("Custom Path", func(t *testing.T) {
		customPath := "/custom/path/config.yaml"
		os.Setenv("HEPHAESTUS_CONFIG_PATH", customPath)
		defer os.Unsetenv("HEPHAESTUS_CONFIG_PATH")

		assert.Equal(t, customPath, GetConfigFilePath())
	})
} 
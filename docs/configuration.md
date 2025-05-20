# Configuration Guide

Hephaestus uses a single configuration system to manage all settings. This document describes how to configure Hephaestus for your needs.

## Configuration File Location

By default, Hephaestus looks for the configuration file at:
- `~/.hephaestus/config.yaml` (Unix-like systems)
- `%USERPROFILE%\.hephaestus\config.yaml` (Windows)

You can override this location by setting the `HEPHAESTUS_CONFIG` environment variable.

## Getting Started

1. Copy the default configuration file from `config/config.yaml` to your configuration directory
2. Update the values marked with "update-with-your-*" with your specific settings
3. Save the file and Hephaestus will use these settings

## Configuration Structure

The configuration file has the following structure:

```yaml
# Remote Repository Settings
remote:
  token: "update-with-your-github-token"      # Required: Your GitHub personal access token
  owner: "update-with-your-org"               # Required: Your GitHub organization or username
  repository: "update-with-your-repo"         # Required: Your repository name
  branch: "main"                             # Optional: Branch to analyze (default: main)

# Model Service Settings
model:
  provider: "openai"                         # Required: AI provider (e.g., openai, anthropic)
  api_key: "update-with-your-api-key"        # Required: Your API key for the model service
  model_version: "gpt-4"                     # Required: Model version to use

# Logging Settings
log:
  level: "info"                              # Optional: Log level (debug, info, warn, error)
  format: "json"                             # Optional: Log format (json, text)

# Repository Settings
repository:
  path: "update-with-your-path"              # Required: Path to your local repository
  max_files: 10000                          # Optional: Max files to process (default: 10000)
  max_file_size: 1048576                    # Optional: Max file size in bytes (default: 1MB)

# Operational Mode
mode: "suggest"                             # Optional: Mode (suggest, deploy)
```

## Configuration Options

### Remote Repository Settings
- `token`: Your GitHub personal access token
- `owner`: Your GitHub organization or username
- `repository`: Your repository name
- `branch`: Branch to analyze (default: "main")

### Model Service Settings
- `provider`: AI provider (e.g., "openai", "anthropic")
- `api_key`: Your API key for the model service
- `model_version`: Model version to use

### Logging Settings
- `level`: Log level (default: "info")
  - Available levels: "debug", "info", "warn", "error"
- `format`: Log format (default: "json")
  - Available formats: "json", "text"

### Repository Settings
- `path`: Path to your local repository
- `max_files`: Maximum number of files to process (default: 10000)
- `max_file_size`: Maximum file size in bytes (default: 1MB)

### Operational Mode
- `mode`: Operation mode (default: "suggest")
  - Available modes: "suggest", "deploy"

## Default Values

If any optional configuration values are not specified, the following defaults will be used:
- `remote.branch`: "main"
- `log.level`: "info"
- `log.format`: "json"
- `repository.max_files`: 10000
- `repository.max_file_size`: 1048576 (1MB)
- `mode`: "suggest"

## Environment Variables

While the configuration file is the preferred method, you can also use environment variables:

- `HEPHAESTUS_CONFIG`: Path to the configuration file
- `HEPHAESTUS_REMOTE_TOKEN`: GitHub authentication token
- `HEPHAESTUS_REMOTE_OWNER`: Repository owner
- `HEPHAESTUS_REMOTE_REPO`: Repository name
- `HEPHAESTUS_REMOTE_BRANCH`: Target branch
- `HEPHAESTUS_MODEL_PROVIDER`: Model service provider
- `HEPHAESTUS_MODEL_API_KEY`: Model service API key
- `HEPHAESTUS_MODEL_VERSION`: Model version
- `HEPHAESTUS_LOG_LEVEL`: Log level
- `HEPHAESTUS_LOG_FORMAT`: Log output format
- `HEPHAESTUS_REPO_PATH`: Repository path
- `HEPHAESTUS_MODE`: Operational mode

Note: Environment variables will override values from the configuration file. 
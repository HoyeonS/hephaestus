# Hephaestus

Hephaestus is a distributed system for automated log analysis and solution generation. It uses machine learning models to analyze system logs and propose solutions for identified issues.

## Features

- Node-based distributed architecture
- Real-time log processing and analysis
- Integration with remote repositories
- Automated solution generation using AI models
- Prometheus metrics collection
- gRPC-based API interface

## Requirements

- Go 1.21 or later
- Prometheus (for metrics collection)
- Access to a remote repository service
- Access to an AI model service provider (OpenAI or Anthropic)

## Installation

```bash
go get github.com/HoyeonS/hephaestus
```

## Configuration

The system requires configuration for:
- Remote repository settings
- Model service settings
- Logging configuration
- Repository settings

Example configuration:

```yaml
remoteSettings:
  authToken: "your-auth-token"
  repositoryOwner: "owner"
  repositoryName: "repo"
  targetBranch: "main"

modelSettings:
  serviceProvider: "openai"
  serviceApiKey: "your-api-key"
  modelVersion: "gpt-4"

loggingSettings:
  logLevel: "info"
  outputFormat: "json"

repositorySettings:
  repositoryPath: "/path/to/repo"
  fileLimit: 1000
  fileSizeLimit: 10485760 # 10MB
```

## Usage

1. Initialize a new node:
```go
node, err := manager.CreateSystemNode(ctx, config)
```

2. Process logs:
```go
err := processor.ProcessLogs(ctx, nodeID, logs)
```

3. Generate solutions:
```go
solution, err := service.GenerateSolution(ctx, nodeID, logEntry)
```

## Metrics

The system collects various metrics using Prometheus:
- Node operations
- Node status
- Log processing duration
- Model latency
- Repository errors

## Development

To run tests:
```bash
go test ./...
```

To build:
```bash
go build ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
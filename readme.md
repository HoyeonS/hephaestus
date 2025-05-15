# Hephaestus

[![Go Reference](https://pkg.go.dev/badge/github.com/HoyeonS/hephaestus.svg)](https://pkg.go.dev/github.com/HoyeonS/hephaestus)
[![Go Report Card](https://goreportcard.com/badge/github.com/HoyeonS/hephaestus)](https://goreportcard.com/report/github.com/HoyeonS/hephaestus)

Hephaestus is an intelligent error detection and automated fix generation system for Go applications. It monitors application logs in real-time, detects critical issues, and either suggests or automatically applies AI-generated fixes.

## Features

- Real-time log monitoring and error detection
- AI-powered error analysis and fix generation
- Automatic code fixes with sandbox testing
- Knowledge base for learning from past fixes
- Support for multiple AI providers (OpenAI, Anthropic, Google)
- Configurable deployment strategies with rollback support

## Installation

```bash
go get github.com/HoyeonS/hephaestus
```

## Quick Start

1. Create a configuration file (`config/config.yaml`):

```yaml
collector:
  log_paths:
    - "/var/log/application/*.log"
  polling_interval: "5s"

generator:
  ai_model:
    provider: "openai"
    model: "gpt-4"
    api_key: "${HEPHAESTUS_AI_KEY}"
    fix_mode: "suggest"
```

2. Initialize Hephaestus in your application:

```go
package main

import (
    "github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
    config, err := loadConfig("config/config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    client, err := hephaestus.New(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Handle suggestions
    go func() {
        for suggestion := range client.GetSuggestionChannel() {
            log.Printf("Fix suggestion: %s\n", suggestion.Description)
        }
    }()
}
```

## Configuration

Hephaestus can be configured through a YAML file with the following sections:

- `collector`: Log collection settings
- `analyzer`: Error analysis configuration
- `generator`: Fix generation and AI model settings
- `deployment`: Deployment and validation settings
- `knowledge`: Knowledge base configuration

See [Configuration Guide](docs/configuration.md) for detailed options.

## Architecture

Hephaestus consists of five main components:

1. **Log Collector**: Monitors log files and streams for errors
2. **Error Analyzer**: Analyzes and classifies detected errors
3. **Fix Generator**: Generates code fixes using AI models
4. **Deployment Service**: Tests and applies fixes
5. **Knowledge Base**: Learns from past fixes

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/HoyeonS/hephaestus.git
cd hephaestus
```

2. Install dependencies:
```bash
go mod download
```

3. Run tests:
```bash
make test
```

### Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for code formatting
- Add comments for exported functions and types
- Write unit tests for new functionality

### Testing

- Write unit tests using the standard `testing` package
- Use table-driven tests where appropriate
- Aim for high test coverage
- Run tests before submitting PRs:
```bash
make test
make lint
```

## Support

- [Documentation](docs/README.md)
- [Issue Tracker](https://github.com/HoyeonS/hephaestus/issues)
- [Discussions](https://github.com/HoyeonS/hephaestus/discussions)

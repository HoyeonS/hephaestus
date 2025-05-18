# Hephaestus

Hephaestus is an intelligent code analysis and automated fix generation service that integrates with your codebase through GitHub. It monitors your application logs, analyzes issues, and either suggests fixes or automatically creates pull requests with solutions.

## Features

- **Automated Code Analysis**: Monitors application logs and identifies issues based on configurable thresholds
- **AI-Powered Fix Generation**: Leverages various AI providers to generate intelligent code fixes
- **GitHub Integration**: Automatically creates pull requests with suggested fixes
- **Flexible Deployment Modes**: 
  - `suggest`: Returns suggested fixes without making changes
  - `deploy`: Automatically creates pull requests with fixes
- **gRPC API**: High-performance API with streaming support for real-time log monitoring

## Getting Started

### Prerequisites

- Go 1.21 or later
- GitHub account and personal access token
- AI provider API key (supported providers: TBD)

### Installation

```bash
go get github.com/HoyeonS/hephaestus
```

### Configuration

Create a configuration file `hephaestus.yaml`:

```yaml
github:
  repository: "owner/repo"
  branch: "main"
  token: "your-github-token"

ai:
  provider: "openai"  # or other supported providers
  api_key: "your-api-key"

log:
  level: "error"  # minimum level to trigger analysis

mode: "suggest"  # or "deploy" for automatic PR creation
```

### Usage

1. Initialize Hephaestus:

```go
import "github.com/HoyeonS/hephaestus"

config := hephaestus.LoadConfig("hephaestus.yaml")
client := hephaestus.NewClient(config)

// Initialize the service
resp, err := client.Initialize(context.Background())
if err != nil {
    log.Fatal(err)
}

// Start listening for logs
stream, err := client.StreamLogs(context.Background())
if err != nil {
    log.Fatal(err)
}

// Handle responses
for {
    fix, err := stream.Recv()
    if err != nil {
        log.Printf("Error receiving fix: %v", err)
        continue
    }
    
    if fix.Mode == "suggest" {
        log.Printf("Suggested fix: %s", fix.Solution)
    } else {
        log.Printf("PR created: %s", fix.PullRequestURL)
    }
}
```

## Architecture

Hephaestus operates as a service that:

1. Initializes with your GitHub repository and configuration
2. Creates a virtual repository representation using FileNodes
3. Listens to application logs via gRPC streaming
4. Analyzes logs when they meet the configured threshold
5. Generates fixes using the configured AI provider
6. Either suggests fixes or creates pull requests based on mode

## API Reference

See [proto/hephaestus.proto](proto/hephaestus.proto) for the complete API specification.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

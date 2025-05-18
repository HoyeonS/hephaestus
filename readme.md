# Hephaestus

Hephaestus is a sophisticated automated system for monitoring, analyzing, and resolving issues in distributed systems. Named after the Greek god of craftsmanship and technology, this tool leverages AI to provide intelligent solutions for system problems.

## Features

- **Automated Log Analysis**: Real-time processing and analysis of system logs using advanced AI models
- **Intelligent Problem Resolution**: AI-powered generation of solution proposals for identified issues
- **Distributed Node Management**: Robust management of distributed system nodes
- **Secure Remote Repository Integration**: Seamless integration with remote repositories for code management
- **Metrics Collection**: Comprehensive system metrics collection and monitoring
- **gRPC-based Communication**: High-performance communication between components using gRPC
- **Flexible Configuration**: Extensive configuration options for all system components

## Architecture

### Core Components

1. **Node Manager** (`internal/node/manager.go`)
   - Handles node registration and lifecycle management
   - Tracks node status and health
   - Manages node configuration

2. **Model Service** (`internal/model/service.go`)
   - Provides AI-powered solution generation
   - Validates proposed solutions
   - Manages model sessions and configurations

3. **Remote Service** (`internal/remote/service.go`)
   - Handles interactions with remote repositories
   - Manages file operations and pull requests
   - Provides repository connection management

4. **Repository Service** (`internal/repository/service.go`)
   - Coordinates repository operations
   - Manages file content retrieval and updates
   - Handles pull request creation

5. **Logger** (`internal/logger/logger.go`)
   - Provides structured logging capabilities
   - Supports multiple output formats
   - Includes context-aware logging

6. **Metrics Collector** (`internal/metrics/collector.go`)
   - Collects system metrics
   - Tracks operation latencies
   - Monitors error rates

### Communication Protocol

The system uses Protocol Buffers and gRPC for efficient communication between components. The protocol definitions can be found in `proto/hephaestus.proto`.

## Getting Started

### Prerequisites

- Go 1.19 or later
- Protocol Buffers compiler (protoc)
- Make
- Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/HoyeonS/hephaestus.git
   cd hephaestus
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Generate Protocol Buffer code:
   ```bash
   make proto
   ```

4. Build the project:
   ```bash
   make build
   ```

### Configuration

Create a configuration file (`config.yaml`) with the following structure:

```yaml
log_level: info
log_output: stdout
node_id: your-node-id

repository:
  owner: your-repo-owner
  name: your-repo-name
  token: your-github-token
  base_path: /path/to/repo
  branch: main

model:
  version: v1
  api_key: your-model-api-key
  base_url: https://api.example.com
  timeout: 30
```

### Running Tests

```bash
make test          # Run all tests
make unit-test     # Run unit tests only
make integration-test  # Run integration tests
```

## Usage

### Starting the Server

```bash
./hephaestus
```

### Using the Client

See `examples/client/main.go` for a complete example of how to use the client:

```go
client := pb.NewHephaestusServiceClient(conn)
```

## Development

### Directory Structure

```
.
├── cmd/            # Command-line applications
├── examples/       # Example code and usage
├── internal/       # Internal packages
│   ├── config/     # Configuration management
│   ├── logger/     # Logging system
│   ├── metrics/    # Metrics collection
│   ├── model/      # AI model service
│   ├── node/       # Node management
│   ├── remote/     # Remote repository service
│   ├── repository/ # Repository management
│   └── server/     # gRPC server implementation
├── pkg/            # Public packages
│   └── hephaestus/ # Core types and interfaces
├── proto/          # Protocol Buffer definitions
└── test/           # Test suites
```

### Making Changes

1. Create a new branch for your changes
2. Make your changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Run linters: `make lint`
6. Submit a pull request

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
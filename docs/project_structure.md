# Project Structure

## Overview

The Hephaestus project follows a clean architecture pattern with clear separation of concerns. Each component is isolated in its own package with well-defined interfaces.

## Directory Structure

```
hephaestus/
├── cmd/                    # Command-line applications
│   └── server/
│       └── main.go        # Server entry point
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   │   ├── config.go     # Configuration structures
│   │   ├── loader.go     # Configuration loading
│   │   ├── validator.go  # Configuration validation
│   │   └── config_test.go
│   ├── node/             # Node management
│   │   ├── manager.go    # Node lifecycle management
│   │   ├── state.go      # Node state machine
│   │   ├── registry.go   # Node registration
│   │   └── node_test.go
│   ├── log/              # Log processing
│   │   ├── processor.go  # Log processing pipeline
│   │   ├── buffer.go     # Log buffering
│   │   ├── filter.go     # Log filtering
│   │   └── log_test.go
│   ├── repository/       # Virtual repository
│   │   ├── manager.go    # Repository management
│   │   ├── cache.go      # File caching
│   │   ├── parser.go     # Code parsing
│   │   └── repo_test.go
│   ├── ai/              # AI integration
│   │   ├── provider.go   # AI provider interface
│   │   ├── context.go    # Context preparation
│   │   ├── solution.go   # Solution generation
│   │   └── ai_test.go
│   ├── github/          # GitHub integration
│   │   ├── client.go     # GitHub client
│   │   ├── pr.go         # Pull request management
│   │   ├── sync.go       # Repository synchronization
│   │   └── github_test.go
│   ├── metrics/         # Metrics collection
│   │   ├── collector.go  # Metrics collection
│   │   ├── exporter.go   # Prometheus export
│   │   └── metrics_test.go
│   └── server/          # gRPC server
│       ├── server.go     # Server implementation
│       ├── handler.go    # Request handlers
│       ├── middleware.go # Server middleware
│       └── server_test.go
├── pkg/                 # Public API packages
│   └── hephaestus/
│       ├── types.go      # Public types
│       ├── client.go     # Client library
│       ├── errors.go     # Error definitions
│       └── api_test.go
├── proto/              # Protocol Buffers
│   └── hephaestus.proto
├── test/              # Integration tests
│   ├── integration/
│   │   └── server_test.go
│   └── performance/
│       └── bench_test.go
└── tools/             # Development tools
    └── codegen/
        └── main.go

```

## Component Responsibilities

### 1. Configuration Management (`internal/config`)
- Loading configuration from files and environment
- Validating configuration values
- Providing typed configuration access
- Managing secrets and credentials

### 2. Node Management (`internal/node`)
- Node lifecycle (create, start, stop, delete)
- State management and transitions
- Resource allocation and cleanup
- Node registry and discovery

### 3. Log Processing (`internal/log`)
- Log ingestion and parsing
- Buffering and batching
- Filtering and routing
- Error detection and classification

### 4. Virtual Repository (`internal/repository`)
- File system abstraction
- Content caching
- Code parsing and analysis
- Change tracking

### 5. AI Integration (`internal/ai`)
- Provider interface implementation
- Context preparation
- Solution generation
- Rate limiting and retries

### 6. GitHub Integration (`internal/github`)
- Repository operations
- Pull request management
- Webhook handling
- Access control

### 7. Metrics Collection (`internal/metrics`)
- Performance metrics
- Health metrics
- Custom metrics
- Prometheus integration

### 8. Server Implementation (`internal/server`)
- gRPC service implementation
- Request handling
- Middleware
- Error handling

## Testing Strategy

### Unit Tests
- Each package has its own `*_test.go` files
- Mock interfaces for external dependencies
- Table-driven tests
- Coverage requirements (>80%)

### Integration Tests
- End-to-end scenarios
- Component interaction tests
- API compatibility tests
- Performance benchmarks

### Test Organization
```go
// Example test file structure
package component

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/golang/mock/gomock"
)

func TestComponent(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Output
        setup    func(*testing.T) *Deps
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Development Guidelines

### Package Design
1. Each package should have:
   - Clear interfaces
   - Minimal dependencies
   - Comprehensive tests
   - Documentation

2. Interface Design:
   ```go
   type Component interface {
       // Methods with clear documentation
       Method(ctx context.Context, param Param) (Result, error)
   }
   ```

3. Error Handling:
   ```go
   // Define package-specific errors
   var (
       ErrNotFound    = errors.New("not found")
       ErrInvalidInput = errors.New("invalid input")
   )
   ```

### Code Organization
1. File Naming:
   - `component.go` - Main implementation
   - `types.go` - Type definitions
   - `errors.go` - Error definitions
   - `component_test.go` - Tests

2. Interface Implementation:
   ```go
   type implementation struct {
       // Dependencies
       dep Dependency
   }

   func New(dep Dependency) Component {
       return &implementation{dep: dep}
   }
   ```

## Build and Test Process

The build process is integrated with testing:
```makefile
.PHONY: build
build: test
    go build ./...

.PHONY: test
test:
    go test -v -race -cover ./...
```

## Metrics and Monitoring

Each component exports metrics:
```go
type Metrics struct {
    Operations    *prometheus.CounterVec
    Latency      *prometheus.HistogramVec
    Errors       *prometheus.CounterVec
}
```

## Logging

Structured logging throughout:
```go
log.With(
    "component", "name",
    "operation", "action",
    "duration", duration,
).Info("Operation completed")
``` 
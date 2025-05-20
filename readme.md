# Hephaestus

A log processing and solution generation system that monitors logs, detects patterns, and generates solutions.

## Features

- **Log Processing**: Real-time log monitoring with configurable thresholds
- **Pattern Detection**: Identifies patterns in log entries
- **Solution Generation**: Generates code changes based on detected patterns
- **Mode-based Operation**: Supports suggest and deploy modes
- **Remote Repository Integration**: Creates pull requests for generated solutions
- **Error Handling**: Comprehensive error handling and status monitoring

## Architecture

### High-Level Design

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Log Entry │────▶│    Node     │────▶│  Solution   │
└─────────────┘     └─────────────┘     └─────────────┘
                           │                   │
                           ▼                   ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  Config     │     │  Remote     │
                    │  Manager    │     │  Repository │
                    └─────────────┘     └─────────────┘
```

### Components

1. **Node**
   - Log buffering and threshold monitoring
   - Solution generation based on log patterns
   - Mode-based solution handling (suggest/deploy)
   - Error handling and status management

2. **Configuration Manager**
   - YAML-based configuration
   - Validation of settings
   - Default value management

3. **Solution Generator**
   - Pattern analysis
   - Code change generation
   - Confidence scoring

4. **Remote Repository Handler**
   - Repository connection management
   - Pull request creation
   - Branch management

## Project Structure

```
.
├── examples/          # Example implementations
│   └── simple/       # Simple example with configuration
├── internal/         # Internal packages
│   ├── node/        # Node implementation
│   ├── config/      # Configuration management
│   ├── logger/      # Logging system
│   ├── repository/  # Repository management
│   ├── remote/      # Remote repository handling
│   ├── metrics/     # Metrics collection
│   ├── log/         # Log processing
│   ├── model/       # Model implementation
│   └── server/      # Server implementation
├── pkg/             # Public packages
│   └── hephaestus/  # Core types and interfaces
├── config/          # Configuration files
├── deployment/      # Deployment configurations
├── proto/          # Protocol definitions
├── Makefile        # Build and development commands
├── Dockerfile      # Container configuration
├── go.mod          # Go module definition
├── go.sum          # Go module checksums
└── .gitignore      # Git ignore rules
```

### Key Directories

1. **examples/**
   - Contains example implementations
   - Simple example with configuration

2. **internal/**
   - `node/`: Node implementation and management
   - `config/`: Configuration loading and validation
   - `logger/`: Logging system implementation
   - `repository/`: Repository management
   - `remote/`: Remote repository integration
   - `metrics/`: Metrics collection and monitoring
   - `log/`: Log processing implementation
   - `model/`: Model implementation
   - `server/`: Server implementation

3. **pkg/**
   - `hephaestus/`: Public types and interfaces
   - Core data structures
   - Public APIs

4. **config/**
   - Configuration templates
   - Default configurations

5. **deployment/**
   - Deployment configurations
   - Container settings

6. **proto/**
   - Protocol definitions
   - API specifications

## Setup Guide

### Prerequisites

- Go 1.19 or later
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/HoyeonS/hephaestus.git
cd hephaestus
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
make build
```

## Configuration

The system is configured using a YAML file. Here's an example configuration:

```yaml
# Log Processing Settings
log:
  threshold_level: "error"    # Level that triggers solution generation
  threshold_count: 3          # Number of threshold logs before triggering
  threshold_window: "5m"      # Time window for threshold counting

# Operation Mode
mode: "suggest"              # suggest or deploy

# Remote Repository Settings (required for deploy mode)
remote_repo:
  token: "your-repo-token"
  owner: "your-org"
  repository: "your-repo"
  branch: "main"
  pr:
    title_template: "fix: {error_type}"
    branch_template: "hephaestus/fix/{error_type}"
    labels: ["hephaestus", "automated"]
```

### Configuration Options

1. **Log Settings**
   - `threshold_level`: Log level to monitor (debug, info, warn, error)
   - `threshold_count`: Number of logs required to trigger processing
   - `threshold_window`: Time window for counting logs

2. **Operation Mode**
   - `suggest`: Only generate and display solutions
   - `deploy`: Generate solutions and create pull requests

3. **Remote Repository Settings**
   - Required only in deploy mode
   - Configures repository connection and PR settings

## Usage Examples

### Basic Usage

1. Create a configuration file (see example above)

2. Initialize and start a node:

```go
manager := config.NewConfigurationManager("hephaestus.yaml")
if err := manager.LoadConfiguration(); err != nil {
    // Handle error
}

node, err := node.NewNode(manager.Get())
if err != nil {
    // Handle error
}

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := node.Start(ctx); err != nil {
    // Handle error
}
```

3. Process logs:

```go
entry := hephaestus.LogEntry{
    Timestamp:   time.Now(),
    Level:       "error",
    Message:     "Your error message",
    Context:     map[string]interface{}{"key": "value"},
    ProcessedAt: time.Now(),
}

if err := node.ProcessLog(entry); err != nil {
    // Handle error
}
```

4. Handle errors:

```go
go func() {
    for err := range node.GetErrors() {
        fmt.Printf("Node error: %v\n", err)
    }
}()
```

### Advanced Usage

1. **Custom Log Processing**:
```go
// Custom log entry with additional context
entry := hephaestus.LogEntry{
    Timestamp:   time.Now(),
    Level:       "error",
    Message:     "Database connection failed",
    Context: map[string]interface{}{
        "database": "main",
        "attempt":  3,
        "error":    "connection timeout",
    },
    ErrorTrace:  "stack trace...",
    ProcessedAt: time.Now(),
}
```

2. **Solution Handling**:
```go
// Custom solution handler
func handleSolution(solution *hephaestus.Solution) error {
    switch solution.Confidence {
    case solution.Confidence > 0.8:
        return deploySolution(solution)
    default:
        return suggestSolution(solution)
    }
}
```

## Error Handling

The system includes comprehensive error handling:

1. **Configuration Errors**
   - Invalid configuration values
   - Missing required settings
   - Invalid mode settings

2. **Processing Errors**
   - Log processing failures
   - Solution generation errors
   - Repository operation errors

3. **Node Status**
   - Initializing
   - Operational
   - Processing
   - Error

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Code Style

- Follow Go standard formatting
- Use meaningful variable names
- Add comments for complex logic
- Write unit tests for new features

## License

MIT

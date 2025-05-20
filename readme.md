# Hephaestus

A log processing and solution generation system.

## Project Structure

```
.
├── examples/          # Example implementations
│   └── simple/       # Simple example with configuration
├── internal/         # Internal packages
│   ├── node/        # Node implementation
│   └── config/      # Configuration management
└── pkg/             # Public packages
    └── hephaestus/  # Core types and interfaces
```

## Core Components

1. **Node**: Processes logs and generates solutions
2. **Configuration**: YAML-based configuration system
3. **Log Processing**: Threshold-based log processing
4. **Solution Generation**: Generates solutions based on log patterns

## Usage

1. Create a configuration file (see `examples/simple/hephaestus.yaml`)
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

## Configuration

See `examples/simple/hephaestus.yaml` for configuration options.

## License

MIT

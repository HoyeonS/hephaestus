# Hephaestus User Guide

## Table of Contents
1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Basic Usage](#basic-usage)
4. [Configuration](#configuration)
5. [Monitoring Options](#monitoring-options)
6. [Error Detection](#error-detection)
7. [Fix Generation](#fix-generation)
8. [Knowledge Base](#knowledge-base)
9. [Metrics and Monitoring](#metrics-and-monitoring)
10. [Best Practices](#best-practices)
11. [Advanced Usage](#advanced-usage)

## Introduction

Hephaestus is an intelligent error detection and automated fix generation system for Go applications. It can be integrated into your application as a library to provide real-time error monitoring, analysis, and automated fix suggestions.

### Key Features
- Real-time log monitoring
- Intelligent error detection
- AI-powered fix generation
- Knowledge base for learning from fixes
- Metrics collection and monitoring
- Multiple AI provider support

## Installation

```bash
go get github.com/yourusername/hephaestus
```

## Basic Usage

Here's a minimal example to get started:

```go
package main

import (
    "github.com/yourusername/hephaestus/pkg/hephaestus"
)

func main() {
    config := hephaestus.DefaultConfig()
    client, _ := hephaestus.NewClient(config)
    client.Start(context.Background())
    defer client.Stop(context.Background())
}
```

## Configuration

### Configuration Options

```go
type Config struct {
    // Log Collection Settings
    LogFormat          string        // "json", "text", or "structured"
    TimeFormat         string        // time format string for parsing timestamps
    ContextTimeWindow  time.Duration // time window for collecting context
    ContextBufferSize  int          // size of the circular buffer for context

    // Error Detection Settings
    ErrorPatterns     map[string]string // error pattern name to regex pattern
    ErrorSeverities   map[string]int    // error pattern name to severity level
    MinErrorSeverity  int              // minimum severity to trigger fix

    // Fix Generation Settings
    MaxFixAttempts    int              // maximum fix attempts per error
    FixTimeout        time.Duration    // timeout for fix generation
    AIProvider        string           // AI provider to use
    AIConfig         map[string]string // AI provider configuration

    // Knowledge Base Settings
    KnowledgeBaseDir  string           // knowledge base directory
    EnableLearning    bool             // enable learning from fixes

    // General Settings
    EnableMetrics     bool             // enable metrics collection
    MetricsEndpoint   string           // metrics endpoint
    LogLevel         string           // log level
}
```

### Detailed Configuration Guide

#### Log Collection Settings

1. **LogFormat** (string)
   - `"json"`: Parse JSON-formatted logs
   - `"text"`: Parse plain text logs
   - `"structured"`: Parse logs with custom delimiters
   
   Example:
   ```go
   config.LogFormat = "json"
   ```

2. **TimeFormat** (string)
   - Use Go time format strings
   - Default: RFC3339
   
   Example:
   ```go
   config.TimeFormat = "2006-01-02T15:04:05Z07:00"
   ```

#### Error Detection Settings

1. **ErrorPatterns** (map[string]string)
   - Define regex patterns for error detection
   - Keys: pattern names
   - Values: regex patterns
   
   Example:
   ```go
   config.ErrorPatterns = map[string]string{
       "panic": `panic:.*`,
       "fatal": `fatal error:.*`,
       "null_pointer": `nil pointer dereference`,
   }
   ```

2. **ErrorSeverities** (map[string]int)
   - Define severity levels for error patterns
   - Levels: 1 (Low) to 4 (Critical)
   
   Example:
   ```go
   config.ErrorSeverities = map[string]int{
       "panic": 4,
       "fatal": 3,
       "null_pointer": 2,
   }
   ```

## Monitoring Options

### Log File Monitoring

```go
file, _ := os.Open("app.log")
client.MonitorReader(ctx, file, "app.log")
```

### Command Output Monitoring

```go
errChan, _ := client.MonitorCommand(ctx, "./myapp", "--debug")
for err := range errChan {
    // Handle errors
}
```

### Custom Reader Monitoring

```go
type CustomReader struct {
    // Your custom reader implementation
}

client.MonitorReader(ctx, customReader, "custom-source")
```

## Error Detection

### Adding Custom Error Patterns

```go
client.AddErrorPattern(`database connection failed: .*`, 3)
```

### Error Context

The system collects context around errors:
- Previous log lines
- System metrics
- Stack traces
- Related variables

## Fix Generation

### AI Provider Configuration

1. **OpenAI Configuration**
   ```go
   config.AIProvider = "openai"
   config.AIConfig["api_key"] = os.Getenv("OPENAI_API_KEY")
   config.AIConfig["model"] = "gpt-4"
   ```

2. **Custom Provider Configuration**
   ```go
   config.AIProvider = "custom"
   config.AIConfig["endpoint"] = "http://your-ai-service"
   ```

### Fix Validation

Fixes are validated through:
1. Syntax checking
2. Test execution
3. Static analysis
4. Human approval (optional)

## Knowledge Base

### Directory Structure
```
hephaestus-kb/
├── errors/
│   ├── <error-hash>/
│   │   ├── context.json
│   │   ├── fix.go
│   │   └── metadata.json
├── patterns/
│   └── patterns.json
└── metrics/
    └── success_rate.json
```

### Learning Mechanism

1. Error Detection
2. Fix Generation
3. Fix Validation
4. Knowledge Storage
5. Pattern Updates

## Metrics and Monitoring

### Available Metrics

```go
type Metrics struct {
    ErrorsDetected    int64
    FixesGenerated    int64
    FixesApplied      int64
    FixesSuccessful   int64
    AverageFixTime    time.Duration
}
```

### Prometheus Integration

```go
config.EnableMetrics = true
config.MetricsEndpoint = ":2112"
```

## Best Practices

1. **Error Pattern Design**
   - Use specific patterns
   - Include context capture
   - Regular validation

2. **Resource Management**
   - Set appropriate timeouts
   - Monitor memory usage
   - Regular cleanup

3. **Security Considerations**
   - Validate fixes
   - Secure AI credentials
   - Access control

## Advanced Usage

### Custom AI Integration

```go
type CustomAI struct {
    // Custom AI implementation
}

func (ai *CustomAI) GenerateFix(ctx context.Context, err *Error) (*Fix, error) {
    // Implementation
}
```

### Custom Fix Validation

```go
type CustomValidator struct {
    // Custom validator implementation
}

func (v *CustomValidator) ValidateFix(fix *Fix) error {
    // Implementation
}
```

### Pipeline Customization

```go
type CustomPipeline struct {
    // Custom pipeline implementation
}

func (p *CustomPipeline) Process(err *Error) (*Fix, error) {
    // Implementation
}
``` 
# Hephaestus Troubleshooting Guide

## Table of Contents
1. [Common Issues](#common-issues)
2. [Error Messages](#error-messages)
3. [Performance Issues](#performance-issues)
4. [Integration Problems](#integration-problems)
5. [AI Provider Issues](#ai-provider-issues)
6. [Knowledge Base Problems](#knowledge-base-problems)
7. [Monitoring Issues](#monitoring-issues)
8. [Debugging Tools](#debugging-tools)

## Common Issues

### Client Initialization Fails

#### Symptoms
- `NewClient` returns an error
- Configuration validation fails
- Client fails to start

#### Solutions
1. Check configuration values:
   ```go
   config := hephaestus.DefaultConfig()
   config.Validate() // Check for specific errors
   ```

2. Verify directory permissions:
   ```bash
   ls -la /path/to/knowledge-base
   chmod 755 /path/to/knowledge-base
   ```

3. Validate AI provider configuration:
   ```go
   if config.AIProvider == "openai" && config.AIConfig["api_key"] == "" {
       log.Fatal("OpenAI API key not configured")
   }
   ```

### Log Monitoring Not Working

#### Symptoms
- No errors detected
- MonitorReader returns immediately
- Missing log entries

#### Solutions
1. Check log format configuration:
   ```go
   config.LogFormat = "json" // Ensure this matches your log format
   ```

2. Verify file permissions:
   ```go
   file, err := os.OpenFile("app.log", os.O_RDONLY, 0644)
   if err != nil {
       log.Printf("Permission error: %v", err)
   }
   ```

3. Monitor with debug logging:
   ```go
   config.LogLevel = "debug"
   ```

## Error Messages

### "Invalid Configuration"

#### Possible Causes
1. Missing required fields
2. Invalid values
3. Incompatible settings

#### Solutions
```go
// Ensure all required fields are set
config := hephaestus.DefaultConfig()
config.LogFormat = "json"
config.TimeFormat = time.RFC3339
config.ContextBufferSize = 1000 // Must be > 0
```

### "Failed to Initialize AI Provider"

#### Possible Causes
1. Invalid API key
2. Network issues
3. Rate limiting

#### Solutions
```go
// Check API key
if key := os.Getenv("OPENAI_API_KEY"); key == "" {
    log.Fatal("OpenAI API key not set")
}

// Test API connection
client.Start(ctx) // Will test connection during initialization
```

## Performance Issues

### High Memory Usage

#### Symptoms
- Increasing memory consumption
- OOM errors
- Slow response times

#### Solutions
1. Adjust buffer sizes:
   ```go
   config.ContextBufferSize = 500 // Reduce from default 1000
   ```

2. Implement cleanup:
   ```go
   // Regular cleanup
   ticker := time.NewTicker(1 * time.Hour)
   go func() {
       for range ticker.C {
           client.Stop(ctx)
           client.Start(ctx)
       }
   }()
   ```

### Slow Fix Generation

#### Symptoms
- Long wait times for fixes
- Timeouts
- Queue buildup

#### Solutions
1. Adjust timeouts:
   ```go
   config.FixTimeout = 60 * time.Second // Increase timeout
   ```

2. Implement retries:
   ```go
   config.MaxFixAttempts = 5 // Increase retry attempts
   ```

## Integration Problems

### Version Compatibility

#### Symptoms
- Build errors
- Runtime panics
- Undefined methods

#### Solutions
1. Check Go version:
   ```bash
   go version
   # Ensure >= go1.16
   ```

2. Update dependencies:
   ```bash
   go get -u github.com/HoyeonS/hephaestus
   go mod tidy
   ```

### Custom Reader Integration

#### Symptoms
- Reader closes unexpectedly
- Missing data
- Parsing errors

#### Solutions
```go
type BufferedReader struct {
    reader io.Reader
    buffer []byte
}

func (br *BufferedReader) Read(p []byte) (n int, err error) {
    // Implement robust reading
}
```

## AI Provider Issues

### OpenAI Integration

#### Symptoms
- Authentication errors
- Rate limiting
- Invalid responses

#### Solutions
1. Check API key:
   ```go
   config.AIConfig["api_key"] = os.Getenv("OPENAI_API_KEY")
   ```

2. Handle rate limits:
   ```go
   config.AIConfig["max_retries"] = "3"
   config.AIConfig["retry_delay"] = "5s"
   ```

### Custom AI Provider

#### Symptoms
- Connection refused
- Timeout errors
- Invalid responses

#### Solutions
```go
type CustomAIProvider struct {
    endpoint string
    client   *http.Client
}

func (p *CustomAIProvider) Connect() error {
    // Implement connection logic
}
```

## Knowledge Base Problems

### Storage Issues

#### Symptoms
- Write failures
- Corrupted data
- Missing entries

#### Solutions
1. Check permissions:
   ```bash
   chmod -R 755 /path/to/knowledge-base
   chown -R user:group /path/to/knowledge-base
   ```

2. Implement backup:
   ```go
   func backupKnowledgeBase(dir string) error {
       // Implement backup logic
   }
   ```

### Learning Issues

#### Symptoms
- No pattern updates
- Invalid learning data
- Performance degradation

#### Solutions
```go
// Implement validation
func validateLearningData(data *LearningData) error {
    // Validation logic
}

// Regular cleanup
func cleanupInvalidData(dir string) error {
    // Cleanup logic
}
```

## Monitoring Issues

### Metrics Collection

#### Symptoms
- Missing metrics
- Invalid values
- Collection stops

#### Solutions
1. Enable debug metrics:
   ```go
   config.EnableMetrics = true
   config.LogLevel = "debug"
   ```

2. Implement health checks:
   ```go
   func checkMetricsHealth(client Client) error {
       metrics, err := client.GetMetrics()
       // Health check logic
   }
   ```

### Prometheus Integration

#### Symptoms
- Endpoint unreachable
- Missing metrics
- Invalid format

#### Solutions
```go
// Configure endpoint
config.MetricsEndpoint = ":2112"
config.EnableMetrics = true

// Implement health check
http.Get("http://localhost:2112/metrics")
```

## Debugging Tools

### Log Analysis

```go
// Enable debug logging
config.LogLevel = "debug"

// Implement log collector
type LogCollector struct {
    logs []string
}

func (lc *LogCollector) Collect(log string) {
    lc.logs = append(lc.logs, log)
}
```

### Metrics Dashboard

```go
// Implement metrics viewer
type MetricsViewer struct {
    client Client
}

func (mv *MetricsViewer) DisplayMetrics() {
    metrics, _ := mv.client.GetMetrics()
    // Display logic
}
```

### Health Checks

```go
// Implement health checker
type HealthChecker struct {
    client Client
}

func (hc *HealthChecker) Check() error {
    // Check client health
    // Check knowledge base
    // Check AI provider
    // Check metrics
    return nil
}
``` 
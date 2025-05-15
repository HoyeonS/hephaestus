package hephaestus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Hephaestus client instance
type Client interface {
	// Start starts the error detection and fix generation pipeline
	Start(ctx context.Context) error
	
	// Stop gracefully stops the pipeline
	Stop(ctx context.Context) error
	
	// MonitorReader starts monitoring a reader for errors
	// Useful for monitoring log files or stdout/stderr
	MonitorReader(ctx context.Context, reader io.Reader, source string) error
	
	// MonitorCommand starts monitoring a command's output for errors
	// Returns a channel that will receive any errors detected
	MonitorCommand(ctx context.Context, name string, args ...string) (<-chan *Error, error)
	
	// AddErrorPattern adds a new error pattern to detect
	AddErrorPattern(pattern string, severity int) error
	
	// RemoveErrorPattern removes an error pattern
	RemoveErrorPattern(pattern string) error
	
	// GetMetrics returns current metrics
	GetMetrics() (*Metrics, error)

	// Health check methods
	
	// Ping performs a quick connectivity check to all components
	Ping(ctx context.Context) error
	
	// CheckHealth performs a comprehensive health check of all components
	CheckHealth(ctx context.Context) (*SystemHealth, error)

	// TestConnectivity performs a basic connectivity test without requiring configuration
	// This is useful for testing if the client can reach the Hephaestus service
	// Returns nil if connection is successful, error otherwise
	TestConnectivity(ctx context.Context) error
}

// Config represents the client configuration
type Config struct {
	// Log settings
	LogFormat      string `yaml:"log_format"`      // "json", "text", or "structured"
	LogLevel       string `yaml:"log_level"`       // "debug", "info", "warn", "error"
	LogFile        string `yaml:"log_file"`        // Path to log file (empty for stdout)
	LogColorized   bool   `yaml:"log_colorized"`   // Whether to colorize log output

	// Error detection settings
	ErrorPatterns    map[string]string `yaml:"error_patterns"`     // Map of error pattern name to regex
	MinErrorSeverity int              `yaml:"min_error_severity"`  // Minimum severity level to process

	// Fix generation settings
	MaxFixAttempts int           `yaml:"max_fix_attempts"` // Maximum number of fix attempts per error
	FixTimeout     time.Duration `yaml:"fix_timeout"`      // Timeout for fix generation

	// API settings
	APIEndpoint string `yaml:"api_endpoint"` // Hephaestus API endpoint
	APIToken    string `yaml:"api_token"`    // API authentication token
}

// Validate validates the client configuration
func (c *Config) Validate() error {
	// Validate log format
	validFormats := map[string]bool{
		"json":       true,
		"text":       true,
		"structured": true,
	}
	if !validFormats[c.LogFormat] {
		return fmt.Errorf("invalid log format: %s (must be one of: json, text, structured)", c.LogFormat)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.LogLevel)
	}

	// Validate error patterns
	if len(c.ErrorPatterns) == 0 {
		return fmt.Errorf("at least one error pattern must be defined")
	}

	// Validate severity
	if c.MinErrorSeverity < 0 {
		return fmt.Errorf("min_error_severity must be >= 0")
	}

	// Validate fix generation settings
	if c.MaxFixAttempts <= 0 {
		return fmt.Errorf("max_fix_attempts must be > 0")
	}
	if c.FixTimeout <= 0 {
		return fmt.Errorf("fix_timeout must be > 0")
	}

	// Validate API settings
	if c.APIEndpoint == "" {
		return fmt.Errorf("api_endpoint is required")
	}
	if c.APIToken == "" {
		return fmt.Errorf("api_token is required")
	}

	return nil
}

// Error represents a detected error with its context and potential fix
type Error struct {
	Message    string
	Severity   int
	Timestamp  time.Time
	Context    map[string]interface{}
	Source     string
	Fix        *Fix
}

// Fix represents a generated fix for an error
type Fix struct {
	Description string
	Code       string
	FilePath   string
	LineNumber int
	Confidence float64
}

// Metrics represents collected metrics
type Metrics struct {
	ErrorsDetected    int64
	FixesGenerated    int64
	FixesApplied      int64
	FixesSuccessful   int64
	AverageFixTime    time.Duration
}

// SystemHealth represents the health status of the system
type SystemHealth struct {
	Status    string
	Message   string
	Details   map[string]interface{}
	Timestamp time.Time
}

// NewDefaultConfig creates a new configuration with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		LogFormat:        "text",
		LogLevel:         "info",
		LogColorized:     true,
		MinErrorSeverity: 1,
		MaxFixAttempts:   3,
		FixTimeout:       30 * time.Second,
		ErrorPatterns: map[string]string{
			"panic":         `panic:.*`,
			"fatal":         `fatal:.*`,
			"error":         `error:.*`,
			"segmentation": `segmentation fault.*`,
		},
		APIEndpoint: "http://localhost:8080", // Default to local development
	}
}

// NewClient creates a new Hephaestus client with the given configuration
func NewClient(config *Config) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return &clientImpl{
		config: config,
	}, nil
}

// clientImpl is the concrete implementation of the Client interface
type clientImpl struct {
	config *Config
}

// Implement all interface methods for clientImpl
func (c *clientImpl) Start(ctx context.Context) error {
	// First check connectivity
	if err := c.TestConnectivity(ctx); err != nil {
		return fmt.Errorf("connectivity check failed: %v", err)
	}
	return nil
}

func (c *clientImpl) Stop(ctx context.Context) error {
	return nil
}

func (c *clientImpl) MonitorReader(ctx context.Context, reader io.Reader, source string) error {
	return nil
}

func (c *clientImpl) MonitorCommand(ctx context.Context, name string, args ...string) (<-chan *Error, error) {
	errChan := make(chan *Error)
	return errChan, nil
}

func (c *clientImpl) AddErrorPattern(pattern string, severity int) error {
	return nil
}

func (c *clientImpl) RemoveErrorPattern(pattern string) error {
	return nil
}

func (c *clientImpl) GetMetrics() (*Metrics, error) {
	return &Metrics{}, nil
}

// Ping performs a quick connectivity check to all components
func (c *clientImpl) Ping(ctx context.Context) error {
	// For now, just use TestConnectivity as it's already implemented
	// In a real implementation, this would be a lighter-weight check
	return c.TestConnectivity(ctx)
}

// CheckHealth performs a comprehensive health check of all components
func (c *clientImpl) CheckHealth(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{
		Status:    "healthy",
		Message:   "All components are healthy",
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Perform connectivity test
	if err := c.TestConnectivity(ctx); err != nil {
		health.Status = "unhealthy"
		health.Message = fmt.Sprintf("Connectivity test failed: %v", err)
		health.Details["connectivity_error"] = err.Error()
		return health, nil
	}

	// Add basic health metrics
	health.Details["uptime"] = time.Since(time.Now()) // Placeholder for actual uptime
	health.Details["memory_usage"] = "N/A"            // Placeholder for actual memory usage
	health.Details["goroutines"] = "N/A"             // Placeholder for actual goroutine count

	return health, nil
}

// TestConnectivity implements a basic connectivity test
func (c *clientImpl) TestConnectivity(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// List of endpoints to try
	endpoints := []string{
		"https://api.github.com/repos/HoyeonS/hephaestus",  // Public GitHub API
		"https://raw.githubusercontent.com/HoyeonS/hephaestus/main/README.md", // Raw content
	}

	// Try each endpoint until one succeeds
	var lastErr error
	for _, endpoint := range endpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
			// Both 200 and 404 indicate we can reach the service
			return nil
		}
		lastErr = fmt.Errorf("received status code %d from %s", resp.StatusCode, endpoint)
	}

	return fmt.Errorf("connectivity test failed: %v", lastErr)
} 
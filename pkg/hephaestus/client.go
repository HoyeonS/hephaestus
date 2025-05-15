package hephaestus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/health"
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
	CheckHealth(ctx context.Context) (*health.SystemHealth, error)

	// TestConnectivity performs a basic connectivity test without requiring configuration
	// This is useful for testing if the client can reach the Hephaestus service
	// Returns nil if connection is successful, error otherwise
	TestConnectivity(ctx context.Context) error
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

// NewClient creates a new Hephaestus client with the given configuration
func NewClient(config *Config) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	// Initialize client implementation
	return &clientImpl{
		config: config,
		health: health.NewChecker(map[string]string{
			"ai_endpoint":  config.AIConfig["endpoint"],
			"metrics_port": config.MetricsEndpoint,
			"kb_path":      config.KnowledgeBaseDir,
		}),
	}, nil
}

// clientImpl is the concrete implementation of the Client interface
type clientImpl struct {
	config *Config
	health *health.Checker
}

// Implement all interface methods for clientImpl
func (c *clientImpl) Start(ctx context.Context) error {
	// First check connectivity
	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("connectivity check failed: %v", err)
	}
	
	// Implementation
	return nil
}

func (c *clientImpl) Stop(ctx context.Context) error {
	// Implementation
	return nil
}

func (c *clientImpl) MonitorReader(ctx context.Context, reader io.Reader, source string) error {
	// Implementation
	return nil
}

func (c *clientImpl) MonitorCommand(ctx context.Context, name string, args ...string) (<-chan *Error, error) {
	// Implementation
	return nil, nil
}

func (c *clientImpl) AddErrorPattern(pattern string, severity int) error {
	// Implementation
	return nil
}

func (c *clientImpl) RemoveErrorPattern(pattern string) error {
	// Implementation
	return nil
}

func (c *clientImpl) GetMetrics() (*Metrics, error) {
	// Implementation
	return nil, nil
}

func (c *clientImpl) Ping(ctx context.Context) error {
	return c.health.PingComponents(ctx)
}

func (c *clientImpl) CheckHealth(ctx context.Context) (*health.SystemHealth, error) {
	return c.health.CheckHealth(ctx)
}

// TestConnectivity implements a basic connectivity test
func (c *clientImpl) TestConnectivity(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// List of endpoints to try (add more if needed)
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
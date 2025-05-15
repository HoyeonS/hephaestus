package hephaestus

import (
	"context"
	"io"
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
	}, nil
}

// clientImpl is the concrete implementation of the Client interface
type clientImpl struct {
	config *Config
	// Add internal fields here
}

// Implement all interface methods for clientImpl
func (c *clientImpl) Start(ctx context.Context) error {
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
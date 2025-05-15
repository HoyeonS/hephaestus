package hephaestus

import (
	"fmt"
	"time"
)

// Config holds all configuration options for Hephaestus
type Config struct {
	// Log collection settings
	LogFormat          string        // "json", "text", or "structured"
	TimeFormat         string        // time format string for parsing timestamps
	ContextTimeWindow  time.Duration // time window for collecting context around errors
	ContextBufferSize  int          // size of the circular buffer for context
	
	// Error detection settings
	ErrorPatterns     map[string]string // map of error pattern name to regex pattern
	ErrorSeverities   map[string]int    // map of error pattern name to severity level
	MinErrorSeverity  int              // minimum severity level to trigger fix generation
	
	// Fix generation settings
	MaxFixAttempts    int              // maximum number of fix attempts per error
	FixTimeout        time.Duration    // timeout for fix generation
	AIProvider        string           // AI provider to use for fix generation
	AIConfig         map[string]string // AI provider specific configuration
	
	// Knowledge base settings
	KnowledgeBaseDir  string           // directory to store knowledge base
	EnableLearning    bool             // whether to enable learning from successful fixes
	
	// Logging settings
	LogLevel         string           // log level (debug, info, warn, error)
	LogColorEnabled  bool             // enable colored log output
	LogComponents    []string         // components to log (empty means all)
	LogFile         string           // log file path (empty means stdout)
	
	// Metrics settings
	EnableMetrics     bool             // whether to collect metrics
	MetricsEndpoint   string           // endpoint for metrics export
	MetricsInterval   time.Duration    // interval for metrics collection
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		// Log collection defaults
		LogFormat:         "json",
		TimeFormat:        time.RFC3339,
		ContextTimeWindow: 5 * time.Minute,
		ContextBufferSize: 1000,
		
		// Error detection defaults
		ErrorPatterns: map[string]string{
			"panic":    `panic:.*`,
			"fatal":    `fatal:.*`,
			"error":    `error:.*`,
		},
		ErrorSeverities: map[string]int{
			"panic": 3, // Critical
			"fatal": 2, // High
			"error": 1, // Medium
		},
		MinErrorSeverity:  1,
		
		// Fix generation defaults
		MaxFixAttempts:    3,
		FixTimeout:        30 * time.Second,
		AIProvider:        "openai",
		AIConfig:          make(map[string]string),
		
		// Knowledge base defaults
		KnowledgeBaseDir:  "./hephaestus-kb",
		EnableLearning:    true,
		
		// Logging defaults
		LogLevel:         "info",
		LogColorEnabled:  true,
		LogComponents:    []string{},
		LogFile:          "",
		
		// Metrics defaults
		EnableMetrics:     false,
		MetricsEndpoint:   ":2112",
		MetricsInterval:   time.Minute,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate log collection settings
	if c.LogFormat != "json" && c.LogFormat != "text" && c.LogFormat != "structured" {
		return fmt.Errorf("invalid log format: %s", c.LogFormat)
	}
	
	if c.ContextTimeWindow <= 0 {
		return fmt.Errorf("context time window must be positive")
	}
	
	if c.ContextBufferSize <= 0 {
		return fmt.Errorf("context buffer size must be positive")
	}
	
	// Validate error detection settings
	if len(c.ErrorPatterns) == 0 {
		return fmt.Errorf("at least one error pattern must be defined")
	}
	
	for name, severity := range c.ErrorSeverities {
		if _, exists := c.ErrorPatterns[name]; !exists {
			return fmt.Errorf("severity defined for non-existent pattern: %s", name)
		}
		if severity < 1 || severity > 3 {
			return fmt.Errorf("invalid severity level for pattern %s: %d", name, severity)
		}
	}
	
	if c.MinErrorSeverity < 1 || c.MinErrorSeverity > 3 {
		return fmt.Errorf("invalid minimum error severity: %d", c.MinErrorSeverity)
	}
	
	// Validate fix generation settings
	if c.MaxFixAttempts <= 0 {
		return fmt.Errorf("max fix attempts must be positive")
	}
	
	if c.FixTimeout <= 0 {
		return fmt.Errorf("fix timeout must be positive")
	}
	
	if c.AIProvider == "" {
		return fmt.Errorf("AI provider must be specified")
	}
	
	// Validate knowledge base settings
	if c.KnowledgeBaseDir == "" {
		return fmt.Errorf("knowledge base directory must be specified")
	}
	
	// Validate logging settings
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	
	// Validate metrics settings
	if c.EnableMetrics {
		if c.MetricsEndpoint == "" {
			return fmt.Errorf("metrics endpoint must be specified when metrics are enabled")
		}
		if c.MetricsInterval <= 0 {
			return fmt.Errorf("metrics interval must be positive")
		}
	}
	
	return nil
} 
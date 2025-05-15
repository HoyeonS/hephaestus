package hephaestus

import (
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
	
	// General settings
	EnableMetrics     bool             // whether to collect metrics
	MetricsEndpoint   string           // endpoint for metrics export
	LogLevel         string           // log level for Hephaestus itself
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LogFormat:         "json",
		TimeFormat:        time.RFC3339,
		ContextTimeWindow: 5 * time.Minute,
		ContextBufferSize: 1000,
		ErrorPatterns: map[string]string{
			"panic":    `panic:.*`,
			"fatal":    `fatal:.*`,
			"error":    `error:.*`,
		},
		ErrorSeverities: map[string]string{
			"panic": "3", // Critical
			"fatal": "2", // High
			"error": "1", // Medium
		},
		MinErrorSeverity:  1,
		MaxFixAttempts:    3,
		FixTimeout:        30 * time.Second,
		AIProvider:        "openai",
		AIConfig:          make(map[string]string),
		KnowledgeBaseDir:  "./hephaestus-kb",
		EnableLearning:    true,
		EnableMetrics:     false,
		MetricsEndpoint:   ":2112",
		LogLevel:          "info",
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Add validation logic here
	return nil
} 
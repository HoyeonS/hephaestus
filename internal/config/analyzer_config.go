package config

import (
	"errors"
	"fmt"
	"time"
)

// AnalyzerConfig holds configuration for the analyzer service
type AnalyzerConfig struct {
	BufferSize       int           `yaml:"buffer_size"`
	ContextWindow    int           `yaml:"context_window"`
	MinErrorSeverity string        `yaml:"min_error_severity"`
	ProcessTimeout   time.Duration `yaml:"process_timeout"`
	MaxRetries      int           `yaml:"max_retries"`
}

// DefaultAnalyzerConfig returns the default configuration for the analyzer
func DefaultAnalyzerConfig() *AnalyzerConfig {
	return &AnalyzerConfig{
		BufferSize:       1000,
		ContextWindow:    5,
		MinErrorSeverity: "low",
		ProcessTimeout:   30 * time.Second,
		MaxRetries:      3,
	}
}

// Validate checks if the analyzer configuration is valid
func (c *AnalyzerConfig) Validate() error {
	if err := validatePositive(c.BufferSize, "buffer size"); err != nil {
		return err
	}
	if err := validateNonNegative(c.ContextWindow, "context window"); err != nil {
		return err
	}
	if err := validateOption(c.MinErrorSeverity, "minimum error severity", validSeverityLevels); err != nil {
		return err
	}
	if c.ProcessTimeout <= 0 {
		return fmt.Errorf("process timeout: %w", ErrNonPositive)
	}
	return validateNonNegative(c.MaxRetries, "max retries")
}

func isValidSeverity(severity string) bool {
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	return validSeverities[severity]
} 
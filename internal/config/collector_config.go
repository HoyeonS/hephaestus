package config

import (
	"fmt"
	"time"
)

// CollectorConfig holds configuration for the collector service
type CollectorConfig struct {
	LogPaths        []string      `yaml:"log_paths"`
	PollingInterval time.Duration `yaml:"polling_interval"`
	BufferSize      int          `yaml:"buffer_size"`
}

// DefaultCollectorConfig returns the default configuration for the collector
func DefaultCollectorConfig() *CollectorConfig {
	return &CollectorConfig{
		LogPaths:        []string{"*.log"},
		PollingInterval: 1 * time.Second,
		BufferSize:      1000,
	}
}

// Validate checks if the collector configuration is valid
func (c *CollectorConfig) Validate() error {
	if len(c.LogPaths) == 0 {
		return fmt.Errorf("log paths: %w", ErrEmptyField)
	}
	if c.PollingInterval <= 0 {
		return fmt.Errorf("polling interval: %w", ErrNonPositive)
	}
	return validatePositive(c.BufferSize, "buffer size")
} 
package config

import (
	"errors"
	"fmt"
	"time"
)

// KnowledgeConfig holds configuration for the knowledge service
type KnowledgeConfig struct {
	StorageType      string        `yaml:"storage_type"`
	StoragePath      string        `yaml:"storage_path"`
	UpdateInterval   time.Duration `yaml:"update_interval"`
	MaxEntries       int           `yaml:"max_entries"`
	RetentionPeriod  time.Duration `yaml:"retention_period"`
	CompressionLevel string        `yaml:"compression_level"`
}

// DefaultKnowledgeConfig returns the default configuration for the knowledge service
func DefaultKnowledgeConfig() *KnowledgeConfig {
	return &KnowledgeConfig{
		StorageType:      "file",
		StoragePath:      "knowledge",
		UpdateInterval:   1 * time.Hour,
		MaxEntries:       10000,
		RetentionPeriod:  30 * 24 * time.Hour, // 30 days
		CompressionLevel: "medium",
	}
}

// Validate checks if the knowledge configuration is valid
func (c *KnowledgeConfig) Validate() error {
	if err := validateOption(c.StorageType, "storage type", validStorageTypes); err != nil {
		return err
	}
	if err := validateNonEmpty(c.StoragePath, "storage path"); err != nil {
		return err
	}
	if c.UpdateInterval <= 0 {
		return fmt.Errorf("update interval: %w", ErrNonPositive)
	}
	if err := validatePositive(c.MaxEntries, "max entries"); err != nil {
		return err
	}
	if c.RetentionPeriod <= 0 {
		return fmt.Errorf("retention period: %w", ErrNonPositive)
	}
	return validateOption(c.CompressionLevel, "compression level", validCompressionLevels)
}

func isValidStorageType(storageType string) bool {
	validTypes := map[string]bool{
		"file":     true,
		"memory":   true,
		"database": true,
	}
	return validTypes[storageType]
}

func isValidCompressionLevel(level string) bool {
	validLevels := map[string]bool{
		"none":   true,
		"low":    true,
		"medium": true,
		"high":   true,
	}
	return validLevels[level]
} 
package config

import (
	"errors"
	"fmt"
	"time"
)

// GeneratorConfig holds configuration for the generator service
type GeneratorConfig struct {
	OutputFormat     string        `yaml:"output_format"`
	TemplateDir      string        `yaml:"template_dir"`
	MaxConcurrent    int           `yaml:"max_concurrent"`
	GenerateTimeout  time.Duration `yaml:"generate_timeout"`
	ValidationLevel  string        `yaml:"validation_level"`
	IncludeMetadata bool          `yaml:"include_metadata"`
}

// DefaultGeneratorConfig returns the default configuration for the generator service
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		OutputFormat:     "yaml",
		TemplateDir:      "templates",
		MaxConcurrent:    5,
		GenerateTimeout:  5 * time.Minute,
		ValidationLevel:  "strict",
		IncludeMetadata: true,
	}
}

// Validate checks if the generator configuration is valid
func (c *GeneratorConfig) Validate() error {
	if err := validateOption(c.OutputFormat, "output format", validFormats); err != nil {
		return err
	}
	if err := validateNonEmpty(c.TemplateDir, "template directory"); err != nil {
		return err
	}
	if err := validatePositive(c.MaxConcurrent, "max concurrent"); err != nil {
		return err
	}
	if c.GenerateTimeout <= 0 {
		return fmt.Errorf("generate timeout: %w", ErrNonPositive)
	}
	return validateOption(c.ValidationLevel, "validation level", validValidationLevels)
}

func isValidFormat(format string) bool {
	validFormats := map[string]bool{
		"yaml": true,
		"json": true,
		"toml": true,
	}
	return validFormats[format]
}

func isValidValidationLevel(level string) bool {
	validLevels := map[string]bool{
		"strict":   true,
		"normal":   true,
		"relaxed":  true,
	}
	return validLevels[level]
} 
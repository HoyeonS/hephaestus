package config

import (
	"errors"
	"fmt"
)

// Common validation errors
var (
	ErrEmptyField     = errors.New("field cannot be empty")
	ErrNonPositive    = errors.New("value must be greater than zero")
	ErrNegative       = errors.New("value must be non-negative")
	ErrInvalidOption  = errors.New("invalid option")
)

// validatePositive checks if a value is greater than zero
func validatePositive(value int, field string) error {
	if value <= 0 {
		return fmt.Errorf("%s: %w", field, ErrNonPositive)
	}
	return nil
}

// validateNonNegative checks if a value is non-negative
func validateNonNegative(value int, field string) error {
	if value < 0 {
		return fmt.Errorf("%s: %w", field, ErrNegative)
	}
	return nil
}

// validateNonEmpty checks if a string is not empty
func validateNonEmpty(value, field string) error {
	if value == "" {
		return fmt.Errorf("%s: %w", field, ErrEmptyField)
	}
	return nil
}

// validateOption checks if a value is in a set of valid options
func validateOption(value, field string, validOptions map[string]bool) error {
	if err := validateNonEmpty(value, field); err != nil {
		return err
	}
	if !validOptions[value] {
		return fmt.Errorf("%s: %w", field, ErrInvalidOption)
	}
	return nil
}

// Common validation maps
var (
	validSeverityLevels = map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}

	validEnvironments = map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	validFormats = map[string]bool{
		"yaml": true,
		"json": true,
		"toml": true,
	}

	validValidationLevels = map[string]bool{
		"strict":  true,
		"normal":  true,
		"relaxed": true,
	}

	validStorageTypes = map[string]bool{
		"file":     true,
		"memory":   true,
		"database": true,
	}

	validCompressionLevels = map[string]bool{
		"none":   true,
		"low":    true,
		"medium": true,
		"high":   true,
	}
) 
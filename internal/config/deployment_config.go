package config

import (
	"errors"
	"fmt"
	"time"
)

// DeploymentConfig holds configuration for the deployment service
type DeploymentConfig struct {
	TargetEnvironment string        `yaml:"target_environment"`
	DeployTimeout     time.Duration `yaml:"deploy_timeout"`
	RetryAttempts     int           `yaml:"retry_attempts"`
	RetryDelay        time.Duration `yaml:"retry_delay"`
	ValidateAfter     bool          `yaml:"validate_after"`
	RollbackEnabled   bool          `yaml:"rollback_enabled"`
}

// DefaultDeploymentConfig returns the default configuration for the deployment service
func DefaultDeploymentConfig() *DeploymentConfig {
	return &DeploymentConfig{
		TargetEnvironment: "development",
		DeployTimeout:     10 * time.Minute,
		RetryAttempts:     3,
		RetryDelay:        30 * time.Second,
		ValidateAfter:     true,
		RollbackEnabled:   true,
	}
}

// Validate checks if the deployment configuration is valid
func (c *DeploymentConfig) Validate() error {
	if err := validateOption(c.TargetEnvironment, "target environment", validEnvironments); err != nil {
		return err
	}
	if c.DeployTimeout <= 0 {
		return fmt.Errorf("deploy timeout: %w", ErrNonPositive)
	}
	if err := validateNonNegative(c.RetryAttempts, "retry attempts"); err != nil {
		return err
	}
	if c.RetryDelay <= 0 {
		return fmt.Errorf("retry delay: %w", ErrNonPositive)
	}
	return nil
}

func isValidEnvironment(env string) bool {
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production": true,
	}
	return validEnvs[env]
} 
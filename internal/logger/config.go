package logger

import (
	"io"
	"os"
)

// Config holds configuration for the logger
type Config struct {
	// Output writer for log messages (defaults to os.Stdout)
	Output io.Writer

	// Prefix for log messages (defaults to "HEPHAESTUS")
	Prefix string

	// MinLevel is the minimum level to log (defaults to INFO)
	MinLevel Level

	// IncludeTimestamp determines if timestamps should be included (defaults to true)
	IncludeTimestamp bool

	// TimeFormat is the format for timestamps (defaults to "2006-01-02 15:04:05.000")
	TimeFormat string
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Output:           os.Stdout,
		Prefix:          "HEPHAESTUS",
		MinLevel:        INFO,
		IncludeTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05.000",
	}
}

// WithOutput sets the output writer
func (c *Config) WithOutput(out io.Writer) *Config {
	c.Output = out
	return c
}

// WithPrefix sets the log prefix
func (c *Config) WithPrefix(prefix string) *Config {
	c.Prefix = prefix
	return c
}

// WithMinLevel sets the minimum log level
func (c *Config) WithMinLevel(level Level) *Config {
	c.MinLevel = level
	return c
}

// WithoutTimestamp disables timestamp logging
func (c *Config) WithoutTimestamp() *Config {
	c.IncludeTimestamp = false
	return c
}

// WithTimeFormat sets the timestamp format
func (c *Config) WithTimeFormat(format string) *Config {
	c.TimeFormat = format
	return c
} 
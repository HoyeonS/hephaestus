package logger

import (
	"fmt"
	"sync"
	"time"
)

// Level represents the severity level of a log message
type Level int

const (
	INFO Level = iota
	WARN
	ERROR
)

// String returns the string representation of a log level
func (l Level) String() string {
	switch l {
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a thread-safe logger instance
type Logger struct {
	config *Config
	mu     sync.Mutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// New creates a new logger with the given configuration
func New(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}
	return &Logger{
		config: config,
	}
}

// Default returns the default logger instance
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New(DefaultConfig())
	})
	return defaultLogger
}

// log writes a message with the given level
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.config.MinLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var parts []string

	if l.config.IncludeTimestamp {
		parts = append(parts, time.Now().Format(l.config.TimeFormat))
	}

	if l.config.Prefix != "" {
		parts = append(parts, l.config.Prefix)
	}

	parts = append(parts, level.String())

	// Build the final message
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	// Construct the log line
	var logLine string
	for _, part := range parts {
		logLine += "[" + part + "] "
	}
	logLine += msg + "\n"

	// Write to output
	fmt.Fprint(l.config.Output, logLine)
}

// Info logs a message at INFO level
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Global convenience functions that use the default logger

// Info logs a message at INFO level using the default logger
func Info(msg string, args ...interface{}) {
	Default().Info(msg, args...)
}

// Warn logs a message at WARN level using the default logger
func Warn(msg string, args ...interface{}) {
	Default().Warn(msg, args...)
}

// Error logs a message at ERROR level using the default logger
func Error(msg string, args ...interface{}) {
	Default().Error(msg, args...)
}

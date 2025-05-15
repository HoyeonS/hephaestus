package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// Level represents the logging level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelColors = map[Level]*color.Color{
	DEBUG: color.New(color.FgCyan),
	INFO:  color.New(color.FgGreen),
	WARN:  color.New(color.FgYellow),
	ERROR: color.New(color.FgRed),
	FATAL: color.New(color.FgRed, color.Bold),
}

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO ",
	WARN:  "WARN ",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger represents a logger instance
type Logger struct {
	level     Level
	component string
	colored   bool
}

// New creates a new logger instance
func New(component string, level Level) *Logger {
	return &Logger{
		level:     level,
		component: component,
		colored:   true,
	}
}

// DisableColor disables colored output
func (l *Logger) DisableColor() {
	l.colored = false
}

// formatMessage formats a log message with timestamp, level, and component
func (l *Logger) formatMessage(level Level, msg string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	formattedMsg := fmt.Sprintf(msg, args...)
	
	baseMsg := fmt.Sprintf("[%s] [%s] [%s] %s",
		timestamp,
		levelNames[level],
		l.component,
		formattedMsg)
	
	if l.colored {
		return levelColors[level].Sprint(baseMsg)
	}
	return baseMsg
}

// log logs a message at the specified level
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	formattedMsg := l.formatMessage(level, msg, args...)
	if level == FATAL {
		fmt.Fprintln(os.Stderr, formattedMsg)
		os.Exit(1)
	} else if level == ERROR {
		fmt.Fprintln(os.Stderr, formattedMsg)
	} else {
		fmt.Println(formattedMsg)
	}
}

// WithComponent creates a new logger with a different component name
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		level:     l.level,
		component: component,
		colored:   l.colored,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(FATAL, msg, args...)
}

// DebugWithFields logs a debug message with structured fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	if DEBUG < l.level {
		return
	}
	l.logWithFields(DEBUG, msg, fields)
}

// InfoWithFields logs an info message with structured fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	if INFO < l.level {
		return
	}
	l.logWithFields(INFO, msg, fields)
}

// WarnWithFields logs a warning message with structured fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	if WARN < l.level {
		return
	}
	l.logWithFields(WARN, msg, fields)
}

// ErrorWithFields logs an error message with structured fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	if ERROR < l.level {
		return
	}
	l.logWithFields(ERROR, msg, fields)
}

// logWithFields logs a message with structured fields
func (l *Logger) logWithFields(level Level, msg string, fields map[string]interface{}) {
	baseMsg := l.formatMessage(level, msg)
	
	// Format fields
	fieldStr := "{"
	for k, v := range fields {
		fieldStr += fmt.Sprintf(" %s=%v", k, v)
	}
	fieldStr += " }"
	
	if l.colored {
		fieldStr = color.New(color.Faint).Sprint(fieldStr)
	}
	
	if level >= ERROR {
		fmt.Fprintln(os.Stderr, baseMsg, fieldStr)
	} else {
		fmt.Println(baseMsg, fieldStr)
	}
} 
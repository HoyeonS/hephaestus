package logger

import (
	"fmt"
	"os"
	"time"
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

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO ",
	WARN:  "WARN ",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger provides structured logging
type Logger struct {
	component string
	level     Level
}

// New creates a new logger for a component
func New(component string, level Level) *Logger {
	return &Logger{
		component: component,
		level:     level,
	}
}

// log prints a log message
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, levelNames[level], l.component, message)

	if level >= ERROR {
		fmt.Fprint(os.Stderr, logLine)
	} else {
		fmt.Print(logLine)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
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
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, levelNames[level], l.component, msg)
	
	// Format fields
	fieldStr := "{"
	for k, v := range fields {
		fieldStr += fmt.Sprintf(" %s=%v", k, v)
	}
	fieldStr += " }\n"
	
	if level >= ERROR {
		fmt.Fprint(os.Stderr, logLine, fieldStr)
	} else {
		fmt.Print(logLine, fieldStr)
	}
} 
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

// Logger provides structured logging with colors
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

// log prints a colored log message
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	switch level {
	case DEBUG:
		color.Cyan("[%s] [DEBUG] [%s] %s", timestamp, l.component, message)
	case INFO:
		color.Green("[%s] [INFO] [%s] %s", timestamp, l.component, message)
	case WARN:
		color.Yellow("[%s] [WARN] [%s] %s", timestamp, l.component, message)
	case ERROR:
		color.Red("[%s] [ERROR] [%s] %s", timestamp, l.component, message)
	case FATAL:
		color.HiRed("[%s] [FATAL] [%s] %s", timestamp, l.component, message)
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
	baseMsg := l.formatMessage(level, msg)
	
	// Format fields
	fieldStr := "{"
	for k, v := range fields {
		fieldStr += fmt.Sprintf(" %s=%v", k, v)
	}
	fieldStr += " }"
	
	if level >= ERROR {
		fmt.Fprintln(os.Stderr, baseMsg, fieldStr)
	} else {
		fmt.Println(baseMsg, fieldStr)
	}
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
	
	if levelColors[level] != nil {
		return levelColors[level].Sprint(baseMsg)
	}
	return baseMsg
} 
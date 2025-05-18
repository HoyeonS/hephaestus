package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// Config represents logger configuration
type Config struct {
	Level       string
	Format      string
	OutputPaths []string
}

// Initialize sets up the global logger with the provided configuration
func Initialize(config Config) error {
	// Validate log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncodingConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Set default format if not specified
	if config.Format == "" {
		config.Format = "json"
	}

	// Create encoder based on format
	var encoder zapcore.Encoder
	switch config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return fmt.Errorf("unsupported output format: %s", config.Format)
	}

	// Set default output paths if not specified
	if len(config.OutputPaths) == 0 {
		config.OutputPaths = []string{"stdout"}
	}

	// Create writers for each output path
	var cores []zapcore.Core
	for _, path := range config.OutputPaths {
		var writer zapcore.WriteSyncer
		switch path {
		case "stdout":
			writer = zapcore.AddSync(os.Stdout)
		case "stderr":
			writer = zapcore.AddSync(os.Stderr)
		default:
			// Create directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %v", err)
			}
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open log file: %v", err)
			}
			writer = zapcore.AddSync(file)
		}
		core := zapcore.NewCore(encoder, writer, level)
		cores = append(cores, core)
	}

	// Create logger with all cores
	logger := zap.New(zapcore.NewTee(cores...))
	globalLogger = logger

	return nil
}

// Debug logs a debug message
func Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		if traceID := GetTraceID(ctx); traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}
		globalLogger.Debug(msg, fields...)
	}
}

// Info logs an info message
func Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		if traceID := GetTraceID(ctx); traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}
		globalLogger.Info(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		if traceID := GetTraceID(ctx); traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}
		globalLogger.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		if traceID := GetTraceID(ctx); traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}
		globalLogger.Error(msg, fields...)
	}
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	if globalLogger == nil {
		return nil
	}
	if traceID := GetTraceID(ctx); traceID != "" {
		return globalLogger.With(zap.String("trace_id", traceID))
	}
	return globalLogger
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// GetTraceID extracts the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
} 
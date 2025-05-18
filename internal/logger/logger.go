package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
)

// Config represents the logger configuration
type Config struct {
	Level      string   `yaml:"level"`
	OutputPath string   `yaml:"output_path"`
}

// Field creates a field for structured logging
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// Initialize sets up the logger with the provided configuration
func Initialize(config *Config) error {
	if config == nil {
		return fmt.Errorf("logger configuration is required")
	}

	// Parse log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	// Create encoder configuration
	encoderConfig := zapcore.EncoderConfig{
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

	// Configure output paths
	var output zapcore.WriteSyncer
	if config.OutputPath == "stdout" {
		output = zapcore.AddSync(os.Stdout)
	} else {
		// Create log directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(config.OutputPath), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// Open log file
		file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		output = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		output,
		level,
	)

	// Create logger
	globalLogger = zap.New(core)

	return nil
}

// Debug logs a message at debug level
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Debug(msg, append(fields, extractTraceID(ctx)...)...)
}

// Info logs a message at info level
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Info(msg, append(fields, extractTraceID(ctx)...)...)
}

// Warn logs a message at warn level
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Warn(msg, append(fields, extractTraceID(ctx)...)...)
}

// Error logs a message at error level
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Error(msg, append(fields, extractTraceID(ctx)...)...)
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.With(extractTraceID(ctx))
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Sync()
}

// Helper functions

// extractTraceID extracts the trace ID from the context
func extractTraceID(ctx context.Context) []zap.Field {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return []zap.Field{zap.String("trace_id", fmt.Sprintf("%v", traceID))}
	}
	return nil
} 
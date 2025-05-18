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

// Config represents the configuration for the logger
type Config struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
	Output string `json:"output" yaml:"output"`
}

// Initialize sets up the global logger with the provided configuration
func Initialize(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("logger configuration is required")
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var output zapcore.WriteSyncer
	if cfg.Output != "" {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(cfg.Output), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// Open log file
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		output = zapcore.AddSync(file)
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, output, level)

	// Create logger
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	if globalLogger == nil {
		return zap.NewNop()
	}

	// Extract trace ID from context if available
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return globalLogger.With(zap.String("trace_id", traceID))
	}

	return globalLogger
}

// Debug logs a message at debug level
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// Info logs a message at info level
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// Warn logs a message at warn level
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// Error logs a message at error level
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Error(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Sync()
} 
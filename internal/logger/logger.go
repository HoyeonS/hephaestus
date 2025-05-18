package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Config represents logger configuration
type Config struct {
	Level       string
	Format      string
	OutputPaths []string
}

// Initialize sets up the global logger with the provided configuration
func Initialize(config Config) error {
	var err error
	once.Do(func() {
		var level zapcore.Level
		if err = level.UnmarshalText([]byte(config.Level)); err != nil {
			return
		}

		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		var encoder zapcore.Encoder
		switch config.Format {
		case "json":
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		case "console":
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		default:
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		}

		// Configure output paths
		var outputs []zapcore.WriteSyncer
		if len(config.OutputPaths) == 0 {
			outputs = append(outputs, zapcore.AddSync(os.Stdout))
		} else {
			for _, path := range config.OutputPaths {
				if path == "stdout" {
					outputs = append(outputs, zapcore.AddSync(os.Stdout))
				} else if path == "stderr" {
					outputs = append(outputs, zapcore.AddSync(os.Stderr))
				} else {
					file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						continue
					}
					outputs = append(outputs, zapcore.AddSync(file))
				}
			}
		}

		core := zapcore.NewCore(
			encoder,
			zapcore.NewMultiWriteSyncer(outputs...),
			level,
		)

		globalLogger = zap.New(core,
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	})

	return err
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	if globalLogger == nil {
		return zap.NewNop()
	}
	return globalLogger.With(zap.String("trace_id", GetTraceID(ctx)))
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

// Debug logs a message at debug level
func Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// Info logs a message at info level
func Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// Warn logs a message at warn level
func Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// Error logs a message at error level
func Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	WithContext(ctx).Error(msg, fields...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1)
func Fatal(ctx context.Context, msg string, fields ...zapcore.Field) {
	WithContext(ctx).Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Sync()
} 
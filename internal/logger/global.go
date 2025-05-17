package logger

var globalLogger *Logger

func init() {
	config := DefaultConfig().
		WithPrefix("HEPHAESTUS LOG")
	
	globalLogger = New(config)
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Global convenience functions
func GlobalInfo(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func GlobalWarn(format string, args ...interface{}) {
	globalLogger.Warn(format, args...)
}

func GlobalError(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
} 
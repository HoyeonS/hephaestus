package log

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// Processor implements the LogProcessingService interface
type Processor struct {
	// Configuration
	config *hephaestus.LoggingConfiguration
	
	// Active log streams
	streams     map[string]*LogStream
	streamMutex sync.RWMutex
}

// LogStream represents an active log processing stream
type LogStream struct {
	NodeID       string
	LogLevel     string
	OutputFormat string
	Buffer       []LogEntry
	LastActivity time.Time
	IsActive     bool
}

// LogEntry represents a processed log entry
type LogEntry struct {
	Timestamp   time.Time
	Level       string
	Message     string
	Context     map[string]interface{}
	StackTrace  string
	ProcessedAt time.Time
}

// NewProcessor creates a new instance of the log processor
func NewProcessor() *Processor {
	return &Processor{
		streams: make(map[string]*LogStream),
	}
}

// Initialize sets up the log processor with the provided configuration
func (p *Processor) Initialize(ctx context.Context, config hephaestus.LoggingConfiguration) error {
	if config.LogLevel == "" {
		return fmt.Errorf("log level is required")
	}

	if !isValidLogLevel(config.LogLevel) {
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	if config.OutputFormat == "" {
		config.OutputFormat = "json" // Default format
	}

	if !isValidOutputFormat(config.OutputFormat) {
		return fmt.Errorf("invalid output format: %s", config.OutputFormat)
	}

	p.config = &config
	return nil
}

// CreateStream creates a new log processing stream for a node
func (p *Processor) CreateStream(ctx context.Context, nodeID string) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	if _, exists := p.streams[nodeID]; exists {
		return fmt.Errorf("stream already exists for node: %s", nodeID)
	}

	stream := &LogStream{
		NodeID:       nodeID,
		LogLevel:     p.config.LogLevel,
		OutputFormat: p.config.OutputFormat,
		Buffer:       make([]LogEntry, 0),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	p.streams[nodeID] = stream
	return nil
}

// ProcessLogs processes a batch of log entries for a node
func (p *Processor) ProcessLogs(ctx context.Context, nodeID string, logs []string) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	stream, exists := p.streams[nodeID]
	if !exists {
		return fmt.Errorf("stream not found for node: %s", nodeID)
	}

	if !stream.IsActive {
		return fmt.Errorf("stream is not active for node: %s", nodeID)
	}

	// Process each log entry
	for _, logStr := range logs {
		entry, err := p.parseLogEntry(logStr)
		if err != nil {
			continue // Skip invalid entries
		}

		// Check if log level meets threshold
		if !isLogLevelMet(entry.Level, stream.LogLevel) {
			continue
		}

		entry.ProcessedAt = time.Now()
		stream.Buffer = append(stream.Buffer, entry)
	}

	stream.LastActivity = time.Now()
	return nil
}

// GetProcessedLogs retrieves processed log entries for a node
func (p *Processor) GetProcessedLogs(ctx context.Context, nodeID string) ([]LogEntry, error) {
	p.streamMutex.RLock()
	defer p.streamMutex.RUnlock()

	stream, exists := p.streams[nodeID]
	if !exists {
		return nil, fmt.Errorf("stream not found for node: %s", nodeID)
	}

	if !stream.IsActive {
		return nil, fmt.Errorf("stream is not active for node: %s", nodeID)
	}

	// Return a copy of the buffer
	logs := make([]LogEntry, len(stream.Buffer))
	copy(logs, stream.Buffer)

	return logs, nil
}

// ClearProcessedLogs clears processed log entries for a node
func (p *Processor) ClearProcessedLogs(ctx context.Context, nodeID string) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	stream, exists := p.streams[nodeID]
	if !exists {
		return fmt.Errorf("stream not found for node: %s", nodeID)
	}

	stream.Buffer = make([]LogEntry, 0)
	return nil
}

// Cleanup removes a log processing stream and its resources
func (p *Processor) Cleanup(ctx context.Context, nodeID string) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	stream, exists := p.streams[nodeID]
	if !exists {
		return fmt.Errorf("stream not found for node: %s", nodeID)
	}

	stream.IsActive = false
	stream.Buffer = nil
	delete(p.streams, nodeID)

	return nil
}

// Helper functions

func (p *Processor) parseLogEntry(logStr string) (LogEntry, error) {
	var entry LogEntry

	// Try parsing as JSON first
	if err := json.Unmarshal([]byte(logStr), &entry); err == nil {
		return entry, nil
	}

	// Fallback to basic parsing
	entry = LogEntry{
		Timestamp: time.Now(),
		Message:   logStr,
		Level:     "info", // Default level
		Context:   make(map[string]interface{}),
	}

	return entry, nil
}

func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	return validLevels[level]
}

func isValidOutputFormat(format string) bool {
	validFormats := map[string]bool{
		"json":     true,
		"text":     true,
		"logfmt":   true,
		"template": true,
	}

	return validFormats[format]
}

func isLogLevelMet(logLevel, thresholdLevel string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	logValue, exists := levels[logLevel]
	if !exists {
		return false
	}

	thresholdValue, exists := levels[thresholdLevel]
	if !exists {
		return false
	}

	return logValue >= thresholdValue
} 
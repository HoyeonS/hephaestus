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

// isValidLogLevel checks if the provided log level is valid
func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	return validLevels[level]
}

// isValidOutputFormat checks if the provided output format is valid
func isValidOutputFormat(format string) bool {
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	return validFormats[format]
}

// CreateStream creates a new log processing stream for a node
func (p *Processor) CreateStream(nodeID string) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	if _, exists := p.streams[nodeID]; exists {
		return fmt.Errorf("stream already exists for node %s", nodeID)
	}

	p.streams[nodeID] = &LogStream{
		NodeID:       nodeID,
		LogLevel:     p.config.LogLevel,
		OutputFormat: p.config.OutputFormat,
		Buffer:       make([]LogEntry, 0),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	return nil
}

// CloseStream closes a log processing stream
func (p *Processor) CloseStream(nodeID string) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	if stream, exists := p.streams[nodeID]; exists {
		stream.IsActive = false
		delete(p.streams, nodeID)
		return nil
	}

	return fmt.Errorf("stream not found for node %s", nodeID)
}

// ProcessLogEntry processes a log entry for a specific node
func (p *Processor) ProcessLogEntry(nodeID string, entry LogEntry) error {
	p.streamMutex.RLock()
	stream, exists := p.streams[nodeID]
	p.streamMutex.RUnlock()

	if !exists {
		return fmt.Errorf("stream not found for node %s", nodeID)
	}

	if !stream.IsActive {
		return fmt.Errorf("stream is not active for node %s", nodeID)
	}

	// Update last activity
	stream.LastActivity = time.Now()

	// Add to buffer
	stream.Buffer = append(stream.Buffer, entry)

	// Process based on output format
	switch stream.OutputFormat {
	case "json":
		return p.processJSONLog(entry)
	case "text":
		return p.processTextLog(entry)
	default:
		return fmt.Errorf("unsupported output format: %s", stream.OutputFormat)
	}
}

// processJSONLog processes a log entry in JSON format
func (p *Processor) processJSONLog(entry LogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	// TODO: Implement actual output handling
	fmt.Println(string(data))
	return nil
}

// processTextLog processes a log entry in text format
func (p *Processor) processTextLog(entry LogEntry) error {
	// TODO: Implement actual output handling
	fmt.Printf("[%s] %s: %s\n", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
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
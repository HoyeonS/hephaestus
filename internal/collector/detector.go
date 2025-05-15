package collector

import (
	"regexp"
	"sync"
	"time"
)

// ErrorSeverity represents the severity level of detected errors
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// DetectedError represents an error found in logs
type DetectedError struct {
	Message   string
	Severity  ErrorSeverity
	Timestamp time.Time
	Context   map[string]interface{}
	Source    string
	Pattern   string
}

// Detector handles error detection in parsed log entries
type Detector struct {
	patterns     map[*regexp.Regexp]ErrorSeverity
	contextLines int
	mu          sync.RWMutex
}

// NewDetector creates a new error detector with the specified patterns
func NewDetector(patterns map[string]ErrorSeverity, contextLines int) (*Detector, error) {
	compiledPatterns := make(map[*regexp.Regexp]ErrorSeverity)
	
	for pattern, severity := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		compiledPatterns[re] = severity
	}

	return &Detector{
		patterns:     compiledPatterns,
		contextLines: contextLines,
	}, nil
}

// DetectError analyzes a parsed log entry for errors
func (d *Detector) DetectError(entry map[string]interface{}, source string) *DetectedError {
	d.mu.RLock()
	defer d.mu.RUnlock()

	message, ok := entry["message"].(string)
	if !ok {
		return nil
	}

	for pattern, severity := range d.patterns {
		if pattern.MatchString(message) {
			timestamp := time.Now()
			if ts, ok := entry["timestamp"].(time.Time); ok {
				timestamp = ts
			}

			return &DetectedError{
				Message:   message,
				Severity: severity,
				Timestamp: timestamp,
				Context:   d.extractContext(entry),
				Source:    source,
				Pattern:   pattern.String(),
			}
		}
	}

	return nil
}

// AddPattern adds a new error pattern with associated severity
func (d *Detector) AddPattern(pattern string, severity ErrorSeverity) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.patterns[re] = severity

	return nil
}

// RemovePattern removes an error pattern
func (d *Detector) RemovePattern(pattern string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for re := range d.patterns {
		if re.String() == pattern {
			delete(d.patterns, re)
			return
		}
	}
}

func (d *Detector) extractContext(entry map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})
	
	// Copy relevant fields to context
	for k, v := range entry {
		if k != "message" && k != "timestamp" {
			context[k] = v
		}
	}

	return context
}

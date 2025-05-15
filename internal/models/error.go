package models

import (
	"time"
)

// Severity represents the severity level of an error
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
)

// Error represents a detected error in the system
type Error struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	Severity    Severity  `json:"severity"`
	StackTrace  string    `json:"stack_trace,omitempty"`
	Source      string    `json:"source"`
	Context     Context   `json:"context"`
	Timestamp   time.Time `json:"timestamp"`
	CodeSnippet string    `json:"code_snippet,omitempty"`
	LineNumber  int       `json:"line_number,omitempty"`
	FileName    string    `json:"file_name,omitempty"`
	Hash        string    `json:"hash"`           // Unique hash of the error pattern
	Fixed       bool      `json:"fixed"`          // Whether this error has been fixed
	FixID       string    `json:"fix_id"`         // Reference to the fix if applied
	RetryCount  int       `json:"retry_count"`    // Number of fix attempts
	LastRetry   time.Time `json:"last_retry"`     // Timestamp of last fix attempt
}

// NewError creates a new Error instance with default values
func NewError(message string, severity Severity, source string) *Error {
	return &Error{
		ID:        generateUUID(),
		Message:   message,
		Severity:  severity,
		Source:    source,
		Timestamp: time.Now(),
		Context:   NewContext(),
	}
}

// SetStackTrace sets the stack trace for the error
func (e *Error) SetStackTrace(stackTrace string) {
	e.StackTrace = stackTrace
}

// SetCodeContext sets the code-related context for the error
func (e *Error) SetCodeContext(fileName string, lineNumber int, snippet string) {
	e.FileName = fileName
	e.LineNumber = lineNumber
	e.CodeSnippet = snippet
}

// MarkAsFixed marks the error as fixed and associates it with a fix
func (e *Error) MarkAsFixed(fixID string) {
	e.Fixed = true
	e.FixID = fixID
}

// IncrementRetryCount increments the retry count and updates the last retry timestamp
func (e *Error) IncrementRetryCount() {
	e.RetryCount++
	e.LastRetry = time.Now()
}

// generateUUID generates a new UUID string
func generateUUID() string {
	// Implementation using a UUID library of your choice
	// For example, using github.com/google/uuid:
	// return uuid.New().String()
	return "implement-uuid-generation"
}

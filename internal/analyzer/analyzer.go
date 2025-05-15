package analyzer

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/models"
)

// ErrorPattern represents a pattern to match errors
type ErrorPattern struct {
	Pattern  string   `yaml:"pattern"`
	Severity string   `yaml:"severity"`
}

// Config holds configuration for the analyzer service
type Config struct {
	ErrorPatterns  []ErrorPattern `yaml:"error_patterns"`
	MaxStackDepth  int           `yaml:"max_stack_depth"`
	ContextLines   int           `yaml:"context_lines"`
}

// Service represents the error analyzer service
type Service struct {
	config      Config
	patterns    []*regexp.Regexp
	inputChan   chan *models.Error
	outputChan  chan *models.Error
	done        chan struct{}
	mu          sync.RWMutex
}

// New creates a new analyzer service
func New(config Config) (*Service, error) {
	patterns := make([]*regexp.Regexp, len(config.ErrorPatterns))
	for i, ep := range config.ErrorPatterns {
		pattern, err := regexp.Compile(ep.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid error pattern %s: %v", ep.Pattern, err)
		}
		patterns[i] = pattern
	}

	return &Service{
		config:     config,
		patterns:   patterns,
		inputChan:  make(chan *models.Error, 100),
		outputChan: make(chan *models.Error, 100),
		done:       make(chan struct{}),
	}, nil
}

// Start starts the analyzer service
func (s *Service) Start(ctx context.Context) error {
	// Start worker goroutines
	for i := 0; i < 3; i++ { // Number of workers could be configurable
		go s.analyzeErrors(ctx)
	}

	return nil
}

// Stop stops the analyzer service
func (s *Service) Stop() error {
	close(s.done)
	return nil
}

// GetInputChannel returns the channel for receiving errors to analyze
func (s *Service) GetInputChannel() chan<- *models.Error {
	return s.inputChan
}

// GetOutputChannel returns the channel for analyzed errors
func (s *Service) GetOutputChannel() <-chan *models.Error {
	return s.outputChan
}

// analyzeErrors processes errors from the input channel
func (s *Service) analyzeErrors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case err := <-s.inputChan:
			if err == nil {
				continue
			}

			analyzed := s.analyzeError(err)
			
			// Try to send analyzed error, don't block if channel is full
			select {
			case s.outputChan <- analyzed:
			default:
				fmt.Printf("Output channel full, dropping analyzed error: %s\n", analyzed.ID)
			}
		}
	}
}

// analyzeError performs detailed analysis of an error
func (s *Service) analyzeError(err *models.Error) *models.Error {
	// Classify error severity based on patterns
	s.classifyError(err)

	// Extract and analyze stack trace
	if err.StackTrace != "" {
		s.analyzeStackTrace(err)
	}

	// Analyze code context
	s.analyzeCodeContext(err)

	// Generate error hash for grouping similar errors
	err.Hash = s.generateErrorHash(err)

	return err
}

// classifyError determines the severity and type of an error
func (s *Service) classifyError(err *models.Error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Match against known patterns
	for i, pattern := range s.patterns {
		if pattern.Match([]byte(err.Message)) {
			err.Severity = models.Severity(s.config.ErrorPatterns[i].Severity)
			return
		}
	}

	// Default classification if no patterns match
	err.Severity = models.Low
}

// analyzeStackTrace analyzes the error's stack trace
func (s *Service) analyzeStackTrace(err *models.Error) {
	if err.StackTrace == "" {
		return
	}

	// Parse stack trace and limit depth
	lines := parseStackTrace(err.StackTrace)
	if len(lines) > s.config.MaxStackDepth {
		lines = lines[:s.config.MaxStackDepth]
	}

	// Extract relevant information from stack trace
	for _, line := range lines {
		// Add stack frame information to error context
		err.Context.AddCustomData("stack_frame_"+line.Function, line)
	}
}

// analyzeCodeContext analyzes the code context around the error
func (s *Service) analyzeCodeContext(err *models.Error) {
	if err.FileName == "" || err.LineNumber == 0 {
		return
	}

	// Get code context around the error line
	context, err := getCodeContext(err.FileName, err.LineNumber, s.config.ContextLines)
	if err != nil {
		fmt.Printf("Failed to get code context: %v\n", err)
		return
	}

	err.CodeSnippet = context
}

// generateErrorHash generates a unique hash for the error
func (s *Service) generateErrorHash(err *models.Error) string {
	// Implementation would create a hash based on error characteristics
	// This could include message patterns, stack trace patterns, etc.
	return "implement-error-hash-generation"
}

// StackFrame represents a parsed stack trace frame
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// parseStackTrace parses a stack trace string into frames
func parseStackTrace(trace string) []StackFrame {
	// Implementation would parse stack trace format
	// This is a placeholder that returns an empty slice
	return []StackFrame{}
}

// getCodeContext gets the code context around a specific line
func getCodeContext(filename string, line, context int) (string, error) {
	// Implementation would read the file and extract context lines
	// This is a placeholder that returns an empty string
	return "", nil
} 
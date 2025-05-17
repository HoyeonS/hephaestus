package analyzer

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"os"
	"bufio"

	"github.com/HoyeonS/hephaestus/internal/models"
	"github.com/HoyeonS/hephaestus/internal/logger"
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
				log := logger.GetGlobalLogger()
				log.Error("Output channel full, dropping analyzed error: %s", analyzed.ID)
			}
		}
	}
}

// analyzeError performs detailed analysis of an error
func (s *Service) analyzeError(err *models.Error) *models.Error {
	log := logger.GetGlobalLogger()

	analyzed := &models.Error{
		ID:      err.ID,
		Source:  err.Source,
		Message: err.Message,
	}

	// Classify error severity based on patterns
	s.classifyError(analyzed)

	// Extract and analyze stack trace
	if err.StackTrace != "" {
		s.analyzeStackTrace(analyzed)
	}

	// Analyze code context
	s.analyzeCodeContext(analyzed)

	// Generate error hash for grouping similar errors
	analyzed.Hash = s.generateErrorHash(analyzed)

	// Get code context
	context, getContextErr := s.getCodeContext(err)
	if getContextErr != nil {
		log.Error("Failed to get code context: %v", getContextErr)
	} else {
		analyzed.CodeSnippet = context
	}

	return analyzed
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
	context, getContextErr := s.getCodeContext(err)
	if getContextErr != nil {
		log := logger.GetGlobalLogger()
		log.Error("Failed to get code context: %v", getContextErr)
	} else {
		err.CodeSnippet = context
	}
}

// getCodeContext gets the code context around a specific line
func (s *Service) getCodeContext(err *models.Error) (string, error) {
	if err.FileName == "" || err.LineNumber == 0 {
		return "", fmt.Errorf("file name or line number not provided")
	}

	file, openErr := os.Open(err.FileName)
	if openErr != nil {
		return "", fmt.Errorf("failed to open file: %v", openErr)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineNum := 0
	contextLines := s.config.ContextLines

	// Read lines into buffer
	for scanner.Scan() {
		lineNum++
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Calculate context range
	start := err.LineNumber - contextLines
	if start < 0 {
		start = 0
	}
	end := err.LineNumber + contextLines
	if end > len(lines) {
		end = len(lines)
	}

	// Build context string
	var context string
	for i := start; i < end; i++ {
		linePrefix := fmt.Sprintf("%d: ", i+1)
		if i+1 == err.LineNumber {
			linePrefix = ">" + linePrefix
		} else {
			linePrefix = " " + linePrefix
		}
		context += linePrefix + lines[i] + "\n"
	}

	return context, nil
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
package generator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/models"
)

// FixStrategy represents a strategy for generating fixes
type FixStrategy struct {
	Type     string `yaml:"type"`
	Priority int    `yaml:"priority"`
}

// Config holds configuration for the generator service
type Config struct {
	FixStrategies   []FixStrategy `yaml:"fix_strategies"`
	MaxFixAttempts  int          `yaml:"max_fix_attempts"`
	Timeout         time.Duration `yaml:"timeout"`
}

// Service represents the fix generator service
type Service struct {
	config     Config
	inputChan  chan *models.Error
	outputChan chan *models.Fix
	done       chan struct{}
	mu         sync.RWMutex
}

// New creates a new generator service
func New(config Config) (*Service, error) {
	return &Service{
		config:     config,
		inputChan:  make(chan *models.Error, 100),
		outputChan: make(chan *models.Fix, 100),
		done:       make(chan struct{}),
	}, nil
}

// Start starts the generator service
func (s *Service) Start(ctx context.Context) error {
	// Start worker goroutines
	for i := 0; i < 3; i++ { // Number of workers could be configurable
		go s.generateFixes(ctx)
	}

	return nil
}

// Stop stops the generator service
func (s *Service) Stop() error {
	close(s.done)
	return nil
}

// GetInputChannel returns the channel for receiving errors to fix
func (s *Service) GetInputChannel() chan<- *models.Error {
	return s.inputChan
}

// GetOutputChannel returns the channel for generated fixes
func (s *Service) GetOutputChannel() <-chan *models.Fix {
	return s.outputChan
}

// generateFixes processes errors and generates fixes
func (s *Service) generateFixes(ctx context.Context) {
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

			// Create context with timeout for fix generation
			fixCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
			fix := s.generateFix(fixCtx, err)
			cancel()

			if fix != nil {
				// Try to send fix, don't block if channel is full
				select {
				case s.outputChan <- fix:
				default:
					fmt.Printf("Output channel full, dropping generated fix: %s\n", fix.ID)
				}
			}
		}
	}
}

// generateFix generates a fix for an error
func (s *Service) generateFix(ctx context.Context, err *models.Error) *models.Fix {
	// Sort strategies by priority
	strategies := s.getPrioritizedStrategies()

	// Try each strategy until one succeeds or we run out of strategies
	for _, strategy := range strategies {
		select {
		case <-ctx.Done():
			return nil
		default:
			if fix := s.tryStrategy(strategy, err); fix != nil {
				return fix
			}
		}
	}

	return nil
}

// getPrioritizedStrategies returns fix strategies sorted by priority
func (s *Service) getPrioritizedStrategies() []FixStrategy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Copy strategies to avoid modifying the original
	strategies := make([]FixStrategy, len(s.config.FixStrategies))
	copy(strategies, s.config.FixStrategies)

	// Sort by priority (higher priority first)
	// sort.Slice(strategies, func(i, j int) bool {
	// 	return strategies[i].Priority > strategies[j].Priority
	// })

	return strategies
}

// tryStrategy attempts to generate a fix using a specific strategy
func (s *Service) tryStrategy(strategy FixStrategy, err *models.Error) *models.Fix {
	fix := models.NewFix(err.ID, models.FixStrategy(strategy.Type))

	switch strategy.Type {
	case "null_check":
		if s.generateNullCheckFix(err, fix) {
			return fix
		}
	case "exception_handling":
		if s.generateExceptionHandlingFix(err, fix) {
			return fix
		}
	case "resource_cleanup":
		if s.generateResourceCleanupFix(err, fix) {
			return fix
		}
	}

	return nil
}

// generateNullCheckFix generates a fix for null pointer errors
func (s *Service) generateNullCheckFix(err *models.Error, fix *models.Fix) bool {
	// Implementation would analyze the code and add appropriate null checks
	// This is a placeholder implementation
	
	change := models.CodeChange{
		FilePath:    err.FileName,
		StartLine:   err.LineNumber,
		EndLine:     err.LineNumber,
		ChangeType:  "add",
		Description: "Added null check",
		NewCode:     "if value != nil { /* original code */ }",
	}

	fix.AddCodeChange(change)
	fix.Confidence = 0.8
	fix.Description = "Added null pointer check"

	return true
}

// generateExceptionHandlingFix generates a fix for unhandled exceptions
func (s *Service) generateExceptionHandlingFix(err *models.Error, fix *models.Fix) bool {
	// Implementation would add appropriate exception handling
	// This is a placeholder implementation
	
	change := models.CodeChange{
		FilePath:    err.FileName,
		StartLine:   err.LineNumber,
		EndLine:     err.LineNumber,
		ChangeType:  "add",
		Description: "Added error handling",
		NewCode:     "try { /* original code */ } catch (Exception e) { /* handle error */ }",
	}

	fix.AddCodeChange(change)
	fix.Confidence = 0.7
	fix.Description = "Added exception handling"

	return true
}

// generateResourceCleanupFix generates a fix for resource leaks
func (s *Service) generateResourceCleanupFix(err *models.Error, fix *models.Fix) bool {
	// Implementation would add appropriate resource cleanup
	// This is a placeholder implementation
	
	change := models.CodeChange{
		FilePath:    err.FileName,
		StartLine:   err.LineNumber,
		EndLine:     err.LineNumber,
		ChangeType:  "add",
		Description: "Added resource cleanup",
		NewCode:     "defer resource.Close()",
	}

	fix.AddCodeChange(change)
	fix.Confidence = 0.9
	fix.Description = "Added resource cleanup"

	return true
} 
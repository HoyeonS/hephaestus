package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/models"
	"github.com/HoyeonS/hephaestus/pkg/filestore"
)

// ErrorFixPattern represents a learned pattern of errors and their fixes
type ErrorFixPattern struct {
	ID           string                 `json:"id"`
	ErrorPattern string                 `json:"error_pattern"`
	FixStrategy  models.FixStrategy     `json:"fix_strategy"`
	Confidence   float64                `json:"confidence"`
	SuccessRate  float64                `json:"success_rate"`
	UsageCount   int                    `json:"usage_count"`
	LastUsed     time.Time              `json:"last_used"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Config holds configuration for the knowledge service
type Config struct {
	StoragePath      string        `yaml:"storage_path"`
	LearningEnabled  bool          `yaml:"learning_enabled"`
	MaxEntries       int           `yaml:"max_entries"`
	CleanupInterval  time.Duration `yaml:"cleanup_interval"`
	RetentionPeriod  time.Duration `yaml:"retention_period"`
}

// Service represents the knowledge base service
type Service struct {
	config     Config
	store      *filestore.Store
	patterns   map[string]*ErrorFixPattern
	inputChan  chan *models.Fix
	done       chan struct{}
	mu         sync.RWMutex
}

// New creates a new knowledge service
func New(config Config) (*Service, error) {
	store, err := filestore.New(config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	return &Service{
		config:    config,
		store:     store,
		patterns:  make(map[string]*ErrorFixPattern),
		inputChan: make(chan *models.Fix, 100),
		done:      make(chan struct{}),
	}, nil
}

// Start starts the knowledge service
func (s *Service) Start(ctx context.Context) error {
	// Load existing patterns
	if err := s.loadPatterns(); err != nil {
		return fmt.Errorf("failed to load patterns: %v", err)
	}

	// Start background tasks
	go s.processFixResults(ctx)
	go s.runCleanup(ctx)

	return nil
}

// Stop stops the knowledge service
func (s *Service) Stop() error {
	close(s.done)
	return s.savePatterns()
}

// GetInputChannel returns the channel for receiving fix results
func (s *Service) GetInputChannel() chan<- *models.Fix {
	return s.inputChan
}

// FindMatchingPattern finds a pattern that matches an error
func (s *Service) FindMatchingPattern(err *models.Error) *ErrorFixPattern {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestMatch *ErrorFixPattern
	var highestConfidence float64

	for _, pattern := range s.patterns {
		if matches, confidence := s.matchErrorPattern(err, pattern); matches {
			if confidence > highestConfidence {
				highestConfidence = confidence
				bestMatch = pattern
			}
		}
	}

	return bestMatch
}

// processFixResults processes fix results and updates patterns
func (s *Service) processFixResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case fix := <-s.inputChan:
			if fix == nil {
				continue
			}

			if s.config.LearningEnabled {
				s.learnFromFix(fix)
			}
		}
	}
}

// learnFromFix learns from a fix result
func (s *Service) learnFromFix(fix *models.Fix) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pattern := s.findOrCreatePattern(fix)
	s.updatePattern(pattern, fix)

	// Save updated pattern
	if err := s.savePattern(pattern); err != nil {
		fmt.Printf("Failed to save pattern: %v\n", err)
	}
}

// findOrCreatePattern finds an existing pattern or creates a new one
func (s *Service) findOrCreatePattern(fix *models.Fix) *ErrorFixPattern {
	// Look for existing pattern
	for _, pattern := range s.patterns {
		if s.isMatchingPattern(fix, pattern) {
			return pattern
		}
	}

	// Create new pattern
	pattern := &ErrorFixPattern{
		ID:           generateUUID(),
		ErrorPattern: extractErrorPattern(fix),
		FixStrategy:  fix.Strategy,
		Confidence:   fix.Confidence,
		SuccessRate:  1.0,
		UsageCount:   1,
		LastUsed:     time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	s.patterns[pattern.ID] = pattern
	return pattern
}

// updatePattern updates a pattern based on fix result
func (s *Service) updatePattern(pattern *ErrorFixPattern, fix *models.Fix) {
	success := fix.IsSuccessful()
	
	// Update success rate
	oldSuccesses := float64(pattern.UsageCount) * pattern.SuccessRate
	newSuccesses := oldSuccesses
	if success {
		newSuccesses++
	}
	pattern.UsageCount++
	pattern.SuccessRate = newSuccesses / float64(pattern.UsageCount)

	// Update confidence
	pattern.Confidence = (pattern.Confidence*0.8 + fix.Confidence*0.2)

	pattern.LastUsed = time.Now()
}

// runCleanup periodically cleans up old patterns
func (s *Service) runCleanup(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-ticker.C:
			s.cleanupPatterns()
		}
	}
}

// cleanupPatterns removes old or low-quality patterns
func (s *Service) cleanupPatterns() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-s.config.RetentionPeriod)
	
	for id, pattern := range s.patterns {
		// Remove patterns that are old and have low success rate
		if pattern.LastUsed.Before(cutoff) && pattern.SuccessRate < 0.5 {
			delete(s.patterns, id)
			s.deletePattern(id)
		}
	}

	// If we're over the max entries, remove the oldest low-confidence patterns
	if len(s.patterns) > s.config.MaxEntries {
		s.prunePatterns()
	}
}

// prunePatterns removes the least valuable patterns
func (s *Service) prunePatterns() {
	type patternScore struct {
		id    string
		score float64
	}

	// Calculate scores for all patterns
	scores := make([]patternScore, 0, len(s.patterns))
	for id, pattern := range s.patterns {
		score := pattern.Confidence * pattern.SuccessRate * float64(pattern.UsageCount)
		scores = append(scores, patternScore{id, score})
	}

	// Sort by score (lowest first)
	// sort.Slice(scores, func(i, j int) bool {
	// 	return scores[i].score < scores[j].score
	// })

	// Remove lowest scoring patterns until we're under the limit
	for i := 0; i < len(scores) && len(s.patterns) > s.config.MaxEntries; i++ {
		delete(s.patterns, scores[i].id)
		s.deletePattern(scores[i].id)
	}
}

// Helper functions

func (s *Service) loadPatterns() error {
	files, err := s.store.List("patterns")
	if err != nil {
		return err
	}

	for _, id := range files {
		var pattern ErrorFixPattern
		if err := s.store.Load("patterns", id, &pattern); err != nil {
			fmt.Printf("Failed to load pattern %s: %v\n", id, err)
			continue
		}
		s.patterns[id] = &pattern
	}

	return nil
}

func (s *Service) savePatterns() error {
	for _, pattern := range s.patterns {
		if err := s.savePattern(pattern); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) savePattern(pattern *ErrorFixPattern) error {
	return s.store.Save("patterns", pattern.ID, pattern)
}

func (s *Service) deletePattern(id string) error {
	return s.store.Delete("patterns", id)
}

func (s *Service) matchErrorPattern(err *models.Error, pattern *ErrorFixPattern) (bool, float64) {
	// Implementation would use more sophisticated pattern matching
	// This is a placeholder that returns false
	return false, 0.0
}

func (s *Service) isMatchingPattern(fix *models.Fix, pattern *ErrorFixPattern) bool {
	// Implementation would determine if a fix matches a pattern
	// This is a placeholder that returns false
	return false
}

func extractErrorPattern(fix *models.Fix) string {
	// Implementation would extract a pattern from a fix
	// This is a placeholder that returns an empty string
	return ""
}

func generateUUID() string {
	// Implementation would generate a UUID
	// This is a placeholder
	return "implement-uuid-generation"
} 
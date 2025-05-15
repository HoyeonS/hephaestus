package collector

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/models"
)

// Config holds configuration for the collector service
type Config struct {
	LogPaths       []string      `yaml:"log_paths"`
	PollingInterval time.Duration `yaml:"polling_interval"`
	BufferSize     int           `yaml:"buffer_size"`
}

// Service represents the log collector service
type Service struct {
	config     Config
	errorChan  chan *models.Error
	done       chan struct{}
	files      map[string]*os.File
	positions  map[string]int64
	mu         sync.RWMutex
}

// New creates a new collector service
func New(config Config) (*Service, error) {
	return &Service{
		config:    config,
		errorChan: make(chan *models.Error, config.BufferSize),
		done:      make(chan struct{}),
		files:     make(map[string]*os.File),
		positions: make(map[string]int64),
	}, nil
}

// Start starts the collector service
func (s *Service) Start(ctx context.Context) error {
	// Initialize file monitoring
	if err := s.initializeFiles(); err != nil {
		return fmt.Errorf("failed to initialize files: %v", err)
	}

	// Start monitoring files
	go s.monitorFiles(ctx)

	return nil
}

// Stop stops the collector service
func (s *Service) Stop() error {
	close(s.done)
	
	// Close all open files
	s.mu.Lock()
	for _, file := range s.files {
		file.Close()
	}
	s.mu.Unlock()

	return nil
}

// GetErrorChannel returns the channel for detected errors
func (s *Service) GetErrorChannel() <-chan *models.Error {
	return s.errorChan
}

// initializeFiles sets up monitoring for all configured log paths
func (s *Service) initializeFiles() error {
	for _, pattern := range s.config.LogPaths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern %s: %v", pattern, err)
		}

		for _, path := range matches {
			if err := s.addFile(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// monitorFiles periodically checks files for changes
func (s *Service) monitorFiles(ctx context.Context) {
	ticker := time.NewTicker(s.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-ticker.C:
			// Check for new files
			for _, pattern := range s.config.LogPaths {
				matches, err := filepath.Glob(pattern)
				if err != nil {
					fmt.Printf("Error matching glob pattern %s: %v\n", pattern, err)
					continue
				}

				for _, path := range matches {
					s.mu.RLock()
					_, exists := s.files[path]
					s.mu.RUnlock()

					if !exists {
						if err := s.addFile(path); err != nil {
							fmt.Printf("Error adding new file %s: %v\n", path, err)
						}
					}
				}
			}

			// Check existing files for changes
			s.mu.RLock()
			paths := make([]string, 0, len(s.files))
			for path := range s.files {
				paths = append(paths, path)
			}
			s.mu.RUnlock()

			for _, path := range paths {
				s.processFile(path)
			}
		}
	}
}

// addFile adds a new file to be monitored
func (s *Service) addFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", path, err)
	}

	// Seek to end of file for new files
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to seek file %s: %v", path, err)
	}

	s.mu.Lock()
	s.files[path] = file
	s.positions[path] = pos
	s.mu.Unlock()

	return nil
}

// processFile reads new content from a file and processes it
func (s *Service) processFile(path string) {
	s.mu.Lock()
	file, exists := s.files[path]
	pos := s.positions[path]
	s.mu.Unlock()

	if !exists {
		return
	}

	// Read new content
	buffer := make([]byte, 4096)
	for {
		n, err := file.ReadAt(buffer, pos)
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			return
		}

		if n == 0 {
			break
		}

		// Process the new content
		s.processContent(path, buffer[:n])
		pos += int64(n)

		if err == io.EOF {
			break
		}
	}

	// Update position
	s.mu.Lock()
	s.positions[path] = pos
	s.mu.Unlock()
}

// processContent analyzes content for errors
func (s *Service) processContent(source string, content []byte) {
	if !containsError(content) {
		return
	}

	// Create and send error event
	errEvent := &models.Error{
		Source:    source,
		Content:   string(content),
		Timestamp: time.Now(),
	}

	select {
	case s.errorChan <- errEvent:
	default:
		// Channel is full, log overflow
		fmt.Printf("Error channel overflow, dropping error from %s\n", source)
	}
}

// containsError checks if content contains error patterns
func containsError(content []byte) bool {
	// Simple error detection - can be enhanced based on requirements
	return true // For now, treat all content as potential errors
}

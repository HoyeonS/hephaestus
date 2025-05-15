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
	"github.com/fsnotify/fsnotify"
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
	watcher    *fsnotify.Watcher
	errorChan  chan *models.Error
	done       chan struct{}
	files      map[string]*os.File
	positions  map[string]int64
	mu         sync.RWMutex
}

// New creates a new collector service
func New(config Config) (*Service, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	return &Service{
		config:    config,
		watcher:   watcher,
		errorChan: make(chan *models.Error, config.BufferSize),
		done:      make(chan struct{}),
		files:     make(map[string]*os.File),
		positions: make(map[string]int64),
	}, nil
}

// Start starts the collector service
func (s *Service) Start(ctx context.Context) error {
	// Initialize file watchers
	if err := s.initializeWatchers(); err != nil {
		return fmt.Errorf("failed to initialize watchers: %v", err)
	}

	// Start watching for file changes
	go s.watchFiles(ctx)

	// Start polling for new files
	go s.pollNewFiles(ctx)

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

	return s.watcher.Close()
}

// GetErrorChannel returns the channel for detected errors
func (s *Service) GetErrorChannel() <-chan *models.Error {
	return s.errorChan
}

// initializeWatchers sets up file watchers for all configured log paths
func (s *Service) initializeWatchers() error {
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

// watchFiles monitors files for changes
func (s *Service) watchFiles(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-s.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				s.processFile(event.Name)
			}
		case err := <-s.watcher.Errors:
			// Log watcher errors but continue watching
			fmt.Printf("Error watching files: %v\n", err)
		}
	}
}

// pollNewFiles periodically checks for new log files
func (s *Service) pollNewFiles(ctx context.Context) {
	ticker := time.NewTicker(s.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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

	return s.watcher.Add(path)
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

	s.mu.Lock()
	s.positions[path] = pos
	s.mu.Unlock()
}

// processContent analyzes log content for errors
func (s *Service) processContent(source string, content []byte) {
	// This is a simple implementation that looks for "ERROR" in logs
	// In a real implementation, this would use more sophisticated error detection
	
	// Example error detection
	if containsError(content) {
		error := &models.Error{
			Message:   string(content),
			Source:    source,
			Severity:  models.High,
			Timestamp: time.Now(),
		}

		// Try to send error, don't block if channel is full
		select {
		case s.errorChan <- error:
		default:
			fmt.Printf("Error channel full, dropping error from %s\n", source)
		}
	}
}

// containsError checks if content contains error indicators
func containsError(content []byte) bool {
	// Simple implementation - in reality, this would be more sophisticated
	return true // Placeholder implementation
}

package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store represents a file-based storage system
type Store struct {
	basePath string
	mu       sync.RWMutex
}

// New creates a new file store instance
func New(basePath string) (*Store, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}

	return &Store{
		basePath: basePath,
	}, nil
}

// Save saves data to a file
func (s *Store) Save(collection string, id string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure collection directory exists
	collectionPath := filepath.Join(s.basePath, collection)
	if err := os.MkdirAll(collectionPath, 0755); err != nil {
		return fmt.Errorf("failed to create collection directory: %v", err)
	}

	// Marshal data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Create file path
	filePath := filepath.Join(collectionPath, fmt.Sprintf("%s.json", id))

	// Write data to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// Load loads data from a file
func (s *Store) Load(collection string, id string, result interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create file path
	filePath := filepath.Join(s.basePath, collection, fmt.Sprintf("%s.json", id))

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("record not found: %s/%s", collection, id)
		}
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Unmarshal data
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return nil
}

// Delete deletes a file
func (s *Store) Delete(collection string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create file path
	filePath := filepath.Join(s.basePath, collection, fmt.Sprintf("%s.json", id))

	// Delete file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("record not found: %s/%s", collection, id)
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// List lists all files in a collection
func (s *Store) List(collection string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create collection path
	collectionPath := filepath.Join(s.basePath, collection)

	// Read directory
	files, err := os.ReadDir(collectionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	// Extract IDs from filenames
	var ids []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if filepath.Ext(name) == ".json" {
			ids = append(ids, name[:len(name)-5]) // Remove .json extension
		}
	}

	return ids, nil
}

// Backup creates a backup of a collection
func (s *Store) Backup(collection string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create backup directory
	backupDir := filepath.Join(s.basePath, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s.tar.gz", collection, timestamp))

	// Create tar.gz archive
	if err := createArchive(filepath.Join(s.basePath, collection), backupFile); err != nil {
		return "", fmt.Errorf("failed to create backup archive: %v", err)
	}

	return backupFile, nil
}

// Restore restores a collection from a backup
func (s *Store) Restore(backupFile string, collection string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "restore_*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract backup archive
	if err := extractArchive(backupFile, tempDir); err != nil {
		return fmt.Errorf("failed to extract backup archive: %v", err)
	}

	// Remove existing collection
	collectionPath := filepath.Join(s.basePath, collection)
	if err := os.RemoveAll(collectionPath); err != nil {
		return fmt.Errorf("failed to remove existing collection: %v", err)
	}

	// Move restored files to collection directory
	if err := os.Rename(filepath.Join(tempDir, collection), collectionPath); err != nil {
		return fmt.Errorf("failed to move restored files: %v", err)
	}

	return nil
}

// Helper functions for archive operations
func createArchive(srcDir, destFile string) error {
	// Implementation for creating tar.gz archive
	return nil
}

func extractArchive(srcFile, destDir string) error {
	// Implementation for extracting tar.gz archive
	return nil
} 
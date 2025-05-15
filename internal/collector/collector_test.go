package collector

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/models"
)

func TestCollector(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "hephaestus-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration
	config := Config{
		LogPaths:        []string{filepath.Join(tempDir, "*.log")},
		PollingInterval: time.Second,
		BufferSize:      10,
	}

	// Create collector service
	collector, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start collector
	if err := collector.Start(ctx); err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}

	// Create test log file
	logFile := filepath.Join(tempDir, "test.log")
	if err := os.WriteFile(logFile, []byte("ERROR: Test error message\n"), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Wait for error to be detected
	select {
	case err := <-collector.GetErrorChannel():
		if err == nil {
			t.Fatal("Received nil error")
		}
		validateError(t, err)
	case <-ctx.Done():
		t.Fatal("Timeout waiting for error")
	}

	// Stop collector
	if err := collector.Stop(); err != nil {
		t.Fatalf("Failed to stop collector: %v", err)
	}
}

func validateError(t *testing.T, err *models.Error) {
	t.Helper()

	if err.Message == "" {
		t.Error("Error message is empty")
	}

	if err.Source == "" {
		t.Error("Error source is empty")
	}

	if err.Severity == "" {
		t.Error("Error severity is empty")
	}

	if err.Timestamp.IsZero() {
		t.Error("Error timestamp is zero")
	}
} 
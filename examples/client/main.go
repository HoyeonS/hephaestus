package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"gopkg.in/yaml.v3"
)

func main() {
	// Load configuration
	config, err := loadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Hephaestus client
	client, err := hephaestus.New(config)
	if err != nil {
		log.Fatalf("Failed to create Hephaestus client: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Hephaestus
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start Hephaestus: %v", err)
	}

	// Handle suggestions and fixes
	go handleSuggestions(client.GetSuggestionChannel())
	go handleFixes(client.GetFixChannel())
	go handleErrors(client.GetErrorChannel())

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := client.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded")
	case <-ctx.Done():
		log.Println("Shutdown complete")
	}
}

func loadConfig(path string) (*hephaestus.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config hephaestus.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

func handleSuggestions(suggestionChan <-chan *hephaestus.FixSuggestion) {
	for suggestion := range suggestionChan {
		log.Printf("[HEPHAESTUS] Fix suggestion for error %s:\n", suggestion.ErrorID)
		log.Printf("Description: %s\n", suggestion.Description)
		log.Printf("Confidence: %.2f\n", suggestion.Confidence)
		log.Printf("Proposed changes:\n")
		for _, change := range suggestion.CodeChanges {
			log.Printf("- File: %s (lines %d-%d)\n", change.FilePath, change.StartLine, change.EndLine)
			log.Printf("  Description: %s\n", change.Description)
			log.Printf("  New code:\n%s\n", change.NewCode)
		}
	}
}

func handleFixes(fixChan <-chan *models.Fix) {
	for fix := range fixChan {
		log.Printf("[HEPHAESTUS] Applied fix for error %s:\n", fix.ErrorID)
		log.Printf("Strategy: %s\n", fix.Strategy)
		log.Printf("Status: %s\n", fix.Status)
		if fix.IsSuccessful() {
			log.Printf("Fix was successfully applied and verified\n")
		} else {
			log.Printf("Fix application failed or was rolled back\n")
		}
	}
}

func handleErrors(errorChan <-chan error) {
	for err := range errorChan {
		log.Printf("[HEPHAESTUS] Error: %v\n", err)
	}
} 
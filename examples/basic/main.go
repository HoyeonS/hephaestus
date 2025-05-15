package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/yourusername/hephaestus/pkg/hephaestus"
)

func main() {
	// Create a configuration with custom settings
	config := hephaestus.DefaultConfig()
	config.LogFormat = "json"
	config.AIProvider = "openai"
	config.AIConfig["api_key"] = os.Getenv("OPENAI_API_KEY")
	
	// Create a new Hephaestus client
	client, err := hephaestus.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	
	// Start the client
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	
	// Monitor a log file
	logFile, err := os.Open("app.log")
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	
	go func() {
		err := client.MonitorReader(ctx, logFile, "app.log")
		if err != nil {
			log.Printf("Error monitoring log file: %v", err)
		}
	}()
	
	// Monitor a command
	errChan, err := client.MonitorCommand(ctx, "./myapp", "--debug")
	if err != nil {
		log.Fatalf("Failed to start command monitoring: %v", err)
	}
	
	// Handle detected errors
	for err := range errChan {
		log.Printf("Detected error: %s", err.Message)
		if err.Fix != nil {
			log.Printf("Suggested fix: %s\n%s", err.Fix.Description, err.Fix.Code)
		}
	}
	
	// Get metrics periodically
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := client.GetMetrics()
			if err != nil {
				log.Printf("Failed to get metrics: %v", err)
				continue
			}
			log.Printf("Metrics: Errors=%d, Fixes=%d, Success=%d",
				metrics.ErrorsDetected,
				metrics.FixesGenerated,
				metrics.FixesSuccessful)
		}
	}
} 
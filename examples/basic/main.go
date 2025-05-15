package main

import (
	"context"
	"os"
	"time"

	"github.com/yourusername/hephaestus/pkg/hephaestus"
	"github.com/yourusername/hephaestus/pkg/logger"
)

func main() {
	// Create a logger for the example
	log := logger.New("example", logger.INFO)

	// Create a configuration with custom settings
	config := hephaestus.DefaultConfig()
	
	// Configure logging
	config.LogLevel = "debug"
	config.LogColorEnabled = true
	config.LogComponents = []string{"collector", "analyzer", "generator"}
	
	// Configure error detection
	config.ErrorPatterns = map[string]string{
		"panic":     `panic:.*`,
		"fatal":     `fatal error:.*`,
		"error":     `error:.*`,
		"nil_ptr":   `nil pointer dereference`,
		"deadlock":  `deadlock detected:.*`,
		"db_error":  `database error:.*`,
	}
	
	config.ErrorSeverities = map[string]int{
		"panic":    3, // Critical
		"fatal":    3,
		"error":    2, // High
		"nil_ptr":  2,
		"deadlock": 2,
		"db_error": 1, // Medium
	}
	
	// Configure AI provider
	config.AIProvider = "openai"
	config.AIConfig["api_key"] = os.Getenv("OPENAI_API_KEY")
	config.AIConfig["model"] = "gpt-4"
	config.AIConfig["temperature"] = "0.7"
	
	// Configure knowledge base
	config.KnowledgeBaseDir = "./data/kb"
	config.EnableLearning = true
	
	// Configure metrics
	config.EnableMetrics = true
	config.MetricsEndpoint = ":2112"
	config.MetricsInterval = 30 * time.Second

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatal("Invalid configuration: %v", err)
	}
	
	// Create a new Hephaestus client
	client, err := hephaestus.NewClient(config)
	if err != nil {
		log.Fatal("Failed to create client: %v", err)
	}
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	
	// Start the client
	log.Info("Starting Hephaestus client...")
	if err := client.Start(ctx); err != nil {
		log.Fatal("Failed to start client: %v", err)
	}
	
	// Monitor a log file
	log.Info("Setting up log file monitoring...")
	logFile, err := os.Open("app.log")
	if err != nil {
		log.Error("Failed to open log file: %v", err)
	} else {
		defer logFile.Close()
		
		go func() {
			log.InfoWithFields("Started log file monitoring", map[string]interface{}{
				"file": "app.log",
			})
			
			err := client.MonitorReader(ctx, logFile, "app.log")
			if err != nil {
				log.ErrorWithFields("Error monitoring log file", map[string]interface{}{
					"error": err,
					"file":  "app.log",
				})
			}
		}()
	}
	
	// Monitor a command
	log.Info("Setting up command monitoring...")
	errChan, err := client.MonitorCommand(ctx, "./myapp", "--debug")
	if err != nil {
		log.Fatal("Failed to start command monitoring: %v", err)
	}
	
	// Handle detected errors
	go func() {
		for err := range errChan {
			log.InfoWithFields("Detected error", map[string]interface{}{
				"message":  err.Message,
				"severity": err.Severity,
				"source":   err.Source,
			})
			
			if err.Fix != nil {
				log.InfoWithFields("Generated fix", map[string]interface{}{
					"description": err.Fix.Description,
					"file":       err.Fix.FilePath,
					"line":       err.Fix.LineNumber,
					"confidence": err.Fix.Confidence,
				})
			}
		}
	}()
	
	// Monitor metrics
	ticker := time.NewTicker(config.MetricsInterval)
	defer ticker.Stop()
	
	log.Info("Starting metrics collection...")
	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down...")
			return
			
		case <-ticker.C:
			metrics, err := client.GetMetrics()
			if err != nil {
				log.Error("Failed to get metrics: %v", err)
				continue
			}
			
			log.InfoWithFields("Metrics update", map[string]interface{}{
				"errors_detected":  metrics.ErrorsDetected,
				"fixes_generated": metrics.FixesGenerated,
				"fixes_applied":   metrics.FixesApplied,
				"fixes_successful": metrics.FixesSuccessful,
				"avg_fix_time":    metrics.AverageFixTime,
			})
		}
	}
} 
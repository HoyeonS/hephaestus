package main

import (
	"context"
	"os"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
	log := logger.GetGlobalLogger()

	// Create client
	client, err := hephaestus.NewClient(nil)
	if err != nil {
		log.Error("Failed to create client: %v", err)
		os.Exit(1)
	}

	// Test connectivity
	ctx := context.Background()
	err = client.TestConnectivity(ctx)

	log.Info("Testing connectivity to Hephaestus...")
	if err != nil {
		log.Error("❌ Connectivity test failed: %v", err)
		os.Exit(1)
	}
	log.Info("✅ Successfully connected to Hephaestus!")

	// Start client
	err = client.Start(ctx)
	if err != nil {
		log.Error("Failed to start client: %v", err)
		os.Exit(1)
	}

	// Open log file
	logFile, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error("Failed to open log file: %v", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Monitor log file
	err = client.MonitorFile(ctx, logFile.Name())
	if err != nil {
		log.Error("Failed to monitor log file: %v", err)
		os.Exit(1)
	}

	// Get metrics periodically
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics, err := client.GetMetrics(ctx)
		if err != nil {
			log.Error("Failed to get metrics: %v", err)
			continue
		}
		log.Info("Current metrics:")
		log.Info("- Errors detected: %d", metrics.ErrorsDetected)
		log.Info("- Fixes generated: %d", metrics.FixesGenerated)
		log.Info("- Fixes applied: %d", metrics.FixesApplied)
		log.Info("- Fixes successful: %d", metrics.FixesSuccessful)
		log.Info("- Average fix time: %v", metrics.AverageFixTime)
	}
} 
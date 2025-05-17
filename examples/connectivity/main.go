package main

import (
	"context"
	"os"

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
} 
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
	// Create a minimal client without any configuration
	client, err := hephaestus.NewClient(hephaestus.DefaultConfig())
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test basic connectivity
	fmt.Println("Testing connectivity to Hephaestus...")
	if err := client.TestConnectivity(ctx); err != nil {
		fmt.Printf("❌ Connectivity test failed: %v\n", err)
		return
	}
	fmt.Println("✅ Successfully connected to Hephaestus!")

	// Start the client
	if err := client.Start(ctx); err != nil {
		fmt.Printf("Failed to start client: %v\n", err)
		return
	}
	defer client.Stop(ctx)

	// Monitor a log file if it exists
	if _, err := os.Stat("app.log"); err == nil {
		file, err := os.Open("app.log")
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
		} else {
			defer file.Close()
			if err := client.MonitorReader(ctx, file, "app.log"); err != nil {
				fmt.Printf("Failed to monitor log file: %v\n", err)
			}
		}
	}

	// Get and display metrics
	metrics, err := client.GetMetrics()
	if err != nil {
		fmt.Printf("Failed to get metrics: %v\n", err)
	} else {
		fmt.Printf("Current metrics:\n")
		fmt.Printf("- Errors detected: %d\n", metrics.ErrorsDetected)
		fmt.Printf("- Fixes generated: %d\n", metrics.FixesGenerated)
		fmt.Printf("- Fixes applied: %d\n", metrics.FixesApplied)
		fmt.Printf("- Fixes successful: %d\n", metrics.FixesSuccessful)
		fmt.Printf("- Average fix time: %v\n", metrics.AverageFixTime)
	}
} 
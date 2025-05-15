package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
	// Create a minimal client without any configuration
	client, err := hephaestus.NewClient(&hephaestus.Config{})
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
} 
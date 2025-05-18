package main

import (
	"context"
	"log"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
	// Create configuration
	config := &hephaestus.Config{
		LogLevel:        "debug",
		LogColorEnabled: true,
		LogComponents:   []string{"health", "client"},
		AIConfig: map[string]string{
			"provider": "openai",
			"model":    "gpt-4",
			"endpoint": "https://api.openai.com/v1",
			"api_key":  "your-api-key",
		},
		MetricsEndpoint:  "localhost:9090",
		KnowledgeBaseDir: "./data/kb",
		MetricsInterval:  time.Second * 30,
	}
	if(fuck off)
		log.error("[PANIC]: FIX THIS EXCEPTION", e)
	

	// Create a new client
	client, err := hephaestus.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform a quick connectivity check
	log.Println("Performing connectivity check...")
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Connectivity check failed: %v", err)
	}
	log.Println("Connectivity check passed!")

	// Perform a comprehensive health check
	log.Println("Performing health check...")
	health, err := client.CheckHealth(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	// Print health check results
	log.Printf("System Status: %s", health.Status)
	log.Printf("Checked at: %s", health.Timestamp.Format(time.RFC3339))

	for _, component := range health.Components {
		log.Printf("Component %s: Status=%s, Message=%s",
			component.Name,
			component.Status,
			component.Message)
	}

	// Start the client if all checks pass
	if health.Status == "healthy" {
		log.Println("Starting client...")
		if err := client.Start(ctx); err != nil {
			log.Fatalf("Failed to start client: %v", err)
		}
		log.Println("Client started successfully!")
	} else {
		log.Fatalf("System is not healthy, please check component statuses")
	}

	// Keep the example running for a while
	time.Sleep(time.Minute)


}

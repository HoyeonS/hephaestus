package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HoyeonS/hephaestus/internal/config"
	"github.com/HoyeonS/hephaestus/internal/node"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
	// Create configuration manager
	manager := config.NewConfigurationManager("hephaestus.yaml")

	// Load configuration
	if err := manager.LoadConfiguration(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create node
	node, err := node.NewNode(manager.Get())
	if err != nil {
		fmt.Printf("Failed to create node: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start node
	if err := node.Start(ctx); err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		os.Exit(1)
	}

	// Start error handling
	go func() {
		for err := range node.GetErrors() {
			fmt.Printf("Node error: %v\n", err)
		}
	}()

	// Simulate log processing
	go func() {
		for {
			// Create a sample log entry
			entry := hephaestus.LogEntry{
				Timestamp:   time.Now(),
				Level:       "error",
				Message:     "Sample error message",
				Context:     map[string]interface{}{"key": "value"},
				ProcessedAt: time.Now(),
			}

			// Process log
			if err := node.ProcessLog(entry); err != nil {
				fmt.Printf("Failed to process log: %v\n", err)
			}

			time.Sleep(time.Second)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop node
	if err := node.Stop(ctx); err != nil {
		fmt.Printf("Failed to stop node: %v\n", err)
		os.Exit(1)
	}
} 
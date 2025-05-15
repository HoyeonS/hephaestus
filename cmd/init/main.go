package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/HoyeonS/hephaestus/internal/analyzer"
	"github.com/HoyeonS/hephaestus/internal/generator"
	"github.com/HoyeonS/hephaestus/internal/deployment"
	"github.com/HoyeonS/hephaestus/internal/knowledge"
	"github.com/HoyeonS/hephaestus/pkg/filestore"
)

var (
	configPath = flag.String("config", "config/config.yaml", "path to configuration file")
)

func main() {
	flag.Parse()

	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Hephaestus...")

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize file store
	store, err := filestore.New(config.StoragePath)
	if err != nil {
		log.Fatalf("Failed to initialize file store: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	collector := collector.New(config.Collector)
	analyzer := analyzer.New(config.Analyzer)
	generator := generator.New(config.Generator)
	deployment := deployment.New(config.Deployment)
	knowledge := knowledge.New(config.Knowledge)

	// Start components
	if err := startComponents(ctx, collector, analyzer, generator, deployment, knowledge); err != nil {
		log.Fatalf("Failed to start components: %v", err)
	}

	// Handle shutdown gracefully
	handleShutdown(ctx, cancel)
}

func loadConfig(path string) (*Config, error) {
	// Implementation to load and parse YAML configuration
	return nil, nil
}

func startComponents(
	ctx context.Context,
	c *collector.Service,
	a *analyzer.Service,
	g *generator.Service,
	d *deployment.Service,
	k *knowledge.Service,
) error {
	// Start collector service
	if err := c.Start(ctx); err != nil {
		return fmt.Errorf("failed to start collector: %v", err)
	}

	// Start analyzer service
	if err := a.Start(ctx); err != nil {
		return fmt.Errorf("failed to start analyzer: %v", err)
	}

	// Start generator service
	if err := g.Start(ctx); err != nil {
		return fmt.Errorf("failed to start generator: %v", err)
	}

	// Start deployment service
	if err := d.Start(ctx); err != nil {
		return fmt.Errorf("failed to start deployment: %v", err)
	}

	// Start knowledge service
	if err := k.Start(ctx); err != nil {
		return fmt.Errorf("failed to start knowledge: %v", err)
	}

	return nil
}

func handleShutdown(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received shutdown signal: %v", sig)

	// Cancel context to notify all components
	cancel()

	// Allow some time for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for components to shut down or timeout
	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded")
	case <-ctx.Done():
		log.Println("All components shut down successfully")
	}
}

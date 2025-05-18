package main

import (
	"context"
	"time"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {
	log := logger.GetGlobalLogger()

	config := hephaestus.DefaultConfig()

	err := config.Validate()

	if err != nil {
		log.Error("ERR : CONFIG %s", err.Error())
		return
	}

	hephaestusConfig,_ := config.HephaestusConfigFactory(
		collector_config.DefaultCollectorConfig(),
		analyzer_config.DefaultAnalyzerConfig(),
		generator.DefaultConfig(),
		deployment.DefaultConfig(),
		knowledge.DefaultConfig(),
	)

	client, err := hephaestus.New(hephaestusConfig)
	if err != nil {
		log.Error("Error occurred: %v", err)
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Start(ctx); err != nil {
		log.Error("Error occurred: %v", err)
	} else {
		log.Info("Hephaestus Initiated...")
	}



	// Block main thread
	// select {}
}

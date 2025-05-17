package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {

	config := hephaestus.DefaultConfig()

	err := config.Validate()

	if err != nil {
		fmt.Println("ERR : CONFIG", err.Error())
	}

	client, err := hephaestus.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Start(ctx); err != nil {
		fmt.Println("Error Occured", err)
	} else {
		fmt.Println("Hephaestus Initated")
	}

}

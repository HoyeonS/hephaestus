package main

import (
	"fmt"
	"log"

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

	client.Start()

}

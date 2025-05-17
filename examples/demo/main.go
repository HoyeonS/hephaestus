package main

import (
	"fmt"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

func main() {

	config := hephaestus.DefaultConfig()

	err := config.Validate()

	if err != nil {
		fmt.Println("ERR : CONFIG", err.Error())
	}

}

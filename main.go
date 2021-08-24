package main

import (
	"github.com/SeraphJACK/HealthCheck/config"
	"github.com/SeraphJACK/HealthCheck/controller"
	"log"
	"os"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Printf("Failed to read configuration: %v\n", err)
		os.Exit(1)
	}

	err = controller.Init()
	if err != nil {
		log.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

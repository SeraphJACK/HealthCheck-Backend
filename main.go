package main

import (
	"flag"
	"github.com/SeraphJACK/HealthCheck/config"
	"github.com/SeraphJACK/HealthCheck/controller"
	"github.com/SeraphJACK/HealthCheck/notify"
	"log"
	"os"
)

func main() {
	flag.Parse()
	err := config.Init()
	if err != nil {
		log.Printf("Failed to read configuration: %v\n", err)
		os.Exit(1)
	}

	err = notify.Init()
	if err != nil {
		log.Printf("Failed to initialize notification: %v\n", err)
		os.Exit(1)
	}

	err = controller.Start()
	if err != nil {
		log.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

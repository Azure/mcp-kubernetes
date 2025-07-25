package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/Azure/mcp-kubernetes/pkg/server"
)

func main() {
	// Create configuration instance and parse command line arguments
	cfg := config.NewConfig()
	if err := cfg.ParseFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Create validator and run validation checks
	v := config.NewValidator(cfg)
	if !v.Validate() {
		fmt.Fprintln(os.Stderr, "Validation failed:")
		v.PrintErrors()
		os.Exit(1)
	}

	// Create and initialize the service
	service := server.NewService(cfg)
	if err := service.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	// Run the service
	if err := service.Run(); err != nil {
		log.Fatalf("Service error: %v\n", err)
	}
}

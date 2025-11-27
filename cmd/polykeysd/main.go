package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xJohnnyboy/polykeys/internal/infrastructure"
)

func main() {
	log.Println("Polykeys daemon starting...")

	// Initialize app
	app, err := infrastructure.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Load configuration
	if err := app.ManageMappingsUC.LoadFromConfig(ctx); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		log.Println("Daemon will run without mappings. Use 'polykeys add' to configure.")
	}

	// Start device monitoring
	if err := app.MonitorDevicesUC.StartMonitoring(ctx); err != nil {
		log.Fatalf("Failed to start monitoring: %v", err)
	}
	defer app.MonitorDevicesUC.StopMonitoring()

	log.Println("Polykeys daemon ready - monitoring for device changes")

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Polykeys daemon stopped")
}

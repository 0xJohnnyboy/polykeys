package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Polykeys daemon starting...")

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

	// TODO: Initialize adapters and use cases
	// TODO: Start device monitoring
	// TODO: Wait for shutdown

	log.Println("Polykeys daemon ready")
	fmt.Println("(Daemon implementation not yet complete)")

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Polykeys daemon stopped")
}

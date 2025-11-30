package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uzdada/protodiff/internal/adapters/bsr"
	"github.com/uzdada/protodiff/internal/adapters/grpc"
	"github.com/uzdada/protodiff/internal/adapters/k8s"
	"github.com/uzdada/protodiff/internal/adapters/web"
	"github.com/uzdada/protodiff/internal/config"
	"github.com/uzdada/protodiff/internal/core/store"
	"github.com/uzdada/protodiff/internal/scanner"
)

func main() {
	log.Println("Starting ProtoDiff - gRPC Schema Drift Monitor")

	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize core components
	dataStore := store.New()

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	log.Println("Kubernetes client initialized")

	// Initialize gRPC reflection client
	grpcClient := grpc.NewReflectionClient()
	log.Println("gRPC reflection client initialized")

	// Initialize BSR client
	var bsrClient bsr.Client
	if os.Getenv("USE_MOCK_BSR") == "true" {
		bsrClient = bsr.NewMockClient()
		log.Println("BSR client initialized (mock mode)")
	} else {
		bsrClient = bsr.NewHTTPClient()
		if os.Getenv("BSR_TOKEN") == "" {
			log.Println("Warning: BSR_TOKEN not set. BSR API calls may fail for private modules.")
		}
		log.Println("BSR client initialized (HTTP mode)")
	}

	// Initialize web server
	webServer, err := web.NewServer(dataStore, cfg.WebAddr)
	if err != nil {
		log.Fatalf("Failed to create web server: %v", err)
	}

	// Initialize scanner
	scannerInstance := scanner.NewScanner(
		k8sClient,
		grpcClient,
		bsrClient,
		dataStore,
		cfg,
	)

	// Setup context and signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start web server in goroutine
	go func() {
		log.Printf("Starting web server on %s", cfg.WebAddr)
		if err := webServer.Start(); err != nil {
			log.Fatalf("Web server error: %v", err)
		}
	}()

	// Start scanner in goroutine
	go func() {
		if err := scannerInstance.Start(ctx); err != nil {
			if err != context.Canceled {
				log.Printf("Scanner error: %v", err)
			}
		}
	}()

	log.Println("ProtoDiff is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, gracefully stopping...")

	// Cancel context to stop scanner
	cancel()

	// Give goroutines time to cleanup
	time.Sleep(2 * time.Second)

	log.Println("ProtoDiff stopped")
}

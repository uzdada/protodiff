// Copyright 2025 ProtoDiff Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command protodiff monitors gRPC schema drift between live Kubernetes pods
// and the Buf Schema Registry (BSR).
//
// It continuously scans gRPC-enabled pods in a Kubernetes cluster, fetches their
// schemas via gRPC reflection, compares them against the canonical schemas stored
// in BSR, and reports any drift through a built-in web dashboard.
//
// Configuration is done through environment variables:
//   - CONFIGMAP_NAMESPACE: Kubernetes namespace for the mapping ConfigMap
//   - CONFIGMAP_NAME: Name of the ConfigMap containing service-to-BSR mappings
//   - DEFAULT_BSR_TEMPLATE: Template for auto-generating BSR module paths
//   - WEB_ADDR: Address for the web dashboard server
//   - SCAN_INTERVAL: Time duration between scan cycles
//   - BSR_TOKEN: Authentication token for BSR API access
//   - USE_MOCK_BSR: Set to "true" to use mock BSR client (for testing)
//
// Example usage:
//
//	export BSR_TOKEN="your-token-here"
//	export WEB_ADDR=":18080"
//	protodiff
//
// The web dashboard will be available at http://localhost:18080
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

const (
	// Graceful shutdown timeout
	gracefulShutdownTimeout = 2 * time.Second
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

	// Initialize BSR client (using BufClient)
	bsrClient := bsr.NewBufClient()
	log.Println("BSR client initialized (buf CLI mode)")

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
	time.Sleep(gracefulShutdownTimeout)

	log.Println("ProtoDiff stopped")
}

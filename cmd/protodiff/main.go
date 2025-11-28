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
	"github.com/uzdada/protodiff/internal/core/store"
	"github.com/uzdada/protodiff/internal/scanner"
)

const (
	// Default configuration values
	defaultConfigMapNamespace = "protodiff-system"
	defaultConfigMapName      = "protodiff-mapping"
	defaultWebAddr            = ":8080"
	defaultScanInterval       = 30 * time.Second
)

func main() {
	log.Println("Starting ProtoDiff - gRPC Schema Drift Monitor")

	// Load configuration from environment variables
	config := loadConfig()

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
	webServer, err := web.NewServer(dataStore, config.WebAddr)
	if err != nil {
		log.Fatalf("Failed to create web server: %v", err)
	}

	// Initialize scanner
	scannerInstance := scanner.NewScanner(
		k8sClient,
		grpcClient,
		bsrClient,
		dataStore,
		scanner.Config{
			ConfigMapNamespace: config.ConfigMapNamespace,
			ConfigMapName:      config.ConfigMapName,
			BSRTemplate:        config.BSRTemplate,
			ScanInterval:       config.ScanInterval,
		},
	)

	// Setup context and signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start web server in goroutine
	go func() {
		log.Printf("Starting web server on %s", config.WebAddr)
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

// appConfig holds the application configuration
type appConfig struct {
	ConfigMapNamespace string
	ConfigMapName      string
	BSRTemplate        string
	WebAddr            string
	ScanInterval       time.Duration
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() appConfig {
	config := appConfig{
		ConfigMapNamespace: getEnv("CONFIGMAP_NAMESPACE", defaultConfigMapNamespace),
		ConfigMapName:      getEnv("CONFIGMAP_NAME", defaultConfigMapName),
		BSRTemplate:        getEnv("DEFAULT_BSR_TEMPLATE", ""),
		WebAddr:            getEnv("WEB_ADDR", defaultWebAddr),
		ScanInterval:       defaultScanInterval,
	}

	// Parse scan interval if provided
	if intervalStr := os.Getenv("SCAN_INTERVAL"); intervalStr != "" {
		if duration, err := time.ParseDuration(intervalStr); err == nil {
			config.ScanInterval = duration
		} else {
			log.Printf("Warning: Invalid SCAN_INTERVAL '%s', using default %s", intervalStr, defaultScanInterval)
		}
	}

	log.Printf("Configuration loaded:")
	log.Printf("  ConfigMap: %s/%s", config.ConfigMapNamespace, config.ConfigMapName)
	log.Printf("  BSR Template: %s", config.BSRTemplate)
	log.Printf("  Web Address: %s", config.WebAddr)
	log.Printf("  Scan Interval: %s", config.ScanInterval)

	return config
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

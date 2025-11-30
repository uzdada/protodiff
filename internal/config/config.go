// Package config provides centralized application configuration management.
//
// Configuration is loaded from environment variables with sensible defaults.
// All components use the same Config struct to ensure consistency.
//
// Environment variables:
//   - CONFIGMAP_NAMESPACE: Namespace for the Kubernetes ConfigMap (default: "protodiff-system")
//   - CONFIGMAP_NAME: Name of the ConfigMap with service mappings (default: "protodiff-mapping")
//   - DEFAULT_BSR_TEMPLATE: Template for BSR module paths like "buf.build/org/{service}"
//   - WEB_ADDR: Web server address (default: ":18080")
//   - SCAN_INTERVAL: Duration between scans (default: "30m")
package config

import (
	"log"
	"os"
	"time"
)

const (
	// Default values
	defaultConfigMapNamespace = "protodiff-system"
	defaultConfigMapName      = "protodiff-mapping"
	defaultWebAddr            = ":18080"
	defaultScanInterval       = 30 * time.Minute

	// Environment variable names
	envConfigMapNamespace = "CONFIGMAP_NAMESPACE"
	envConfigMapName      = "CONFIGMAP_NAME"
	envBSRTemplate        = "DEFAULT_BSR_TEMPLATE"
	envWebAddr            = "WEB_ADDR"
	envScanInterval       = "SCAN_INTERVAL"
)

// Config holds the application configuration
type Config struct {
	// Kubernetes ConfigMap settings
	ConfigMapNamespace string
	ConfigMapName      string

	// BSR (Buf Schema Registry) settings
	BSRTemplate string

	// Scanner settings
	ScanInterval time.Duration

	// Web server settings
	WebAddr string
}

// Load loads configuration from environment variables with defaults
func Load() Config {
	config := Config{
		ConfigMapNamespace: getEnv(envConfigMapNamespace, defaultConfigMapNamespace),
		ConfigMapName:      getEnv(envConfigMapName, defaultConfigMapName),
		BSRTemplate:        getEnv(envBSRTemplate, ""),
		WebAddr:            getEnv(envWebAddr, defaultWebAddr),
		ScanInterval:       defaultScanInterval,
	}

	// Parse scan interval if provided
	if intervalStr := os.Getenv(envScanInterval); intervalStr != "" {
		if duration, err := time.ParseDuration(intervalStr); err == nil {
			config.ScanInterval = duration
		} else {
			log.Printf("Warning: Invalid SCAN_INTERVAL '%s', using default %s", intervalStr, config.ScanInterval)
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

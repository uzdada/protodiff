package config

import (
	"log"
	"os"
	"time"
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
		ConfigMapNamespace: getEnv("CONFIGMAP_NAMESPACE", "protodiff-system"),
		ConfigMapName:      getEnv("CONFIGMAP_NAME", "protodiff-mapping"),
		BSRTemplate:        getEnv("DEFAULT_BSR_TEMPLATE", ""),
		WebAddr:            getEnv("WEB_ADDR", ":18080"),
		ScanInterval:       30 * time.Minute, // Default scan interval
	}

	// Parse scan interval if provided
	if intervalStr := os.Getenv("SCAN_INTERVAL"); intervalStr != "" {
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

package config

import (
	"os"
)

// Config holds the configuration for the cache node.
type Config struct {
	NodeID   string
	HTTPPort string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		NodeID:   getEnv("CACHE_NODE_ID", "node-1"),
		HTTPPort: getEnv("CACHE_HTTP_PORT", "7001"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

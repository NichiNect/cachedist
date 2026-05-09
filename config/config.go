package config

import (
	"os"
	"strconv"
)

// Config holds the configuration for the cache node.
type Config struct {
	NodeID      string
	HTTPPort    string
	GRPCPort    string
	Peers       string
	NumShards   int
	MaxItems    int
	TTLCleanup  int
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		NodeID:     getEnv("CACHE_NODE_ID", "node-1"),
		HTTPPort:   getEnv("CACHE_HTTP_PORT", "7001"),
		GRPCPort:   getEnv("CACHE_GRPC_PORT", "8001"),
		Peers:      getEnv("CACHE_PEERS", ""),
		NumShards:  getEnvAsInt("CACHE_NUM_SHARDS", 256),
		MaxItems:   getEnvAsInt("CACHE_MAX_ITEMS", 1000000),
		TTLCleanup: getEnvAsInt("CACHE_TTL_CLEANUP", 30),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return fallback
	}
	return value
}

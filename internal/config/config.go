package config

import (
	"os"
)

type Config struct {
	GinMode   string
	HTTPPort  string
	RedisAddr string
}

// Load reads configuration from environment variables with sensible defaults
func Load() Config {
	return Config{
		GinMode:   getEnv("GIN_MODE", "debug"),
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

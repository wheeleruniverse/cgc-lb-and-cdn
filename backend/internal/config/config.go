package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Server ServerConfig `json:"server"`
	Images ImagesConfig `json:"images"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// ImagesConfig holds image-related configuration
type ImagesConfig struct {
	Directory string `json:"directory"`
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvOrDefault("PORT", "8080"),
			Host: getEnvOrDefault("HOST", "0.0.0.0"),
		},
		Images: ImagesConfig{
			Directory: getEnvOrDefault("IMAGES_DIR", "images"),
		},
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

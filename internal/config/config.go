package config

import (
	"log"
	"os"
)

// Config holds the application configuration
type Config struct {
	DatabaseDSN string
	ServerPort  string
}

// LoadConfig reads configuration from environment variables
func LoadConfig() *Config {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		// Provide a default fallback based on .env.example
		log.Println("Warning: DATABASE_DSN environment variable not set. Using default.")
		dsn = "postgres://postgres:postgres_secure_pass@localhost:5433/workout_tracker?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseDSN: dsn,
		ServerPort:  port,
	}
}

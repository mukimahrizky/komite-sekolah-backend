package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	
	// Security
	JWTSecret string
	
	// Server
	ServerPort string
	Environment string // "development" or "production"
	
	// CORS
	AllowedOrigins string // Comma-separated origins for CORS
}

var AppConfig *Config

func Load() {
	// Load .env file (ignore error if file doesn't exist)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	AppConfig = &Config{
		// Database
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "komite_sekolah"),
		
		// Security
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		
		// Server
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		
		// CORS - defaults to localhost for development, MUST be set in production
		// Default: http://localhost:3000 (frontend dev server)
		// In production, you MUST specify exact origins (e.g., "https://yourdomain.com,https://www.yourdomain.com")
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Port            string
	GinMode         string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	JWTExpiryHours  int
	RateLimitRPM    int
	AllowedOrigins  []string
}

// Load reads configuration from environment variables
// Panics if required environment variables are missing
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		DatabaseURL:     getEnvRequired("DATABASE_URL"),
		RedisURL:        getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:       getEnvRequired("JWT_SECRET"),
		JWTExpiryHours:  getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		RateLimitRPM:    getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		AllowedOrigins:  getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

// getEnvRequired gets an environment variable or panics if not set
// Use this for critical configuration that should prevent app startup if missing
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("FATAL: Required environment variable " + key + " is not set. Application cannot start.")
	}
	return value
}

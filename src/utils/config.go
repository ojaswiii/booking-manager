package utils

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	// Server configuration
	ServerPort string
	ServerHost string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Application configuration
	Environment string
	LogLevel    string

	// Booking configuration
	BookingExpiryMinutes int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		// Server configuration
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "localhost"),

		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "ojaswi"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "ticket_booking"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis configuration
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Application configuration
		Environment: getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		// Booking configuration
		BookingExpiryMinutes: getEnvAsInt("BOOKING_EXPIRY_MINUTES", 15),
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDBConnectionString returns the database connection string
func (c *Config) GetDBConnectionString() string {
	// Use URL format for more reliable connection
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)

	return connStr
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

// IsProduction returns true if environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

package services

import (
	"os"
	"strconv"
)

// getEnvInt retrieves an integer environment variable with a fallback default.
// Returns the parsed integer value if the env var exists and is a valid integer.
// Otherwise returns the provided defaultValue.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

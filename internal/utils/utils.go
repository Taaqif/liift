// Package utils provides utility functions for internal use.
package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue == "" {
		log.Printf("ERROR: Required environment variable '%s' is not set and no default value provided", key)
		panic(fmt.Sprintf("required environment variable '%s' is not set", key))
	}
	return defaultValue
}

// GetEnvAsInt gets an environment variable as integer or returns a default value
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

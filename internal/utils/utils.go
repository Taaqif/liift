// Package utils provides utility functions for internal use.
package utils

import (
	"crypto/rand"
	"encoding/hex"
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

func Map[T any, R any](in []T, fn func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

// Set is a set of comparable values (e.g. IDs). Uses map[T]struct{} for zero allocation.
type Set[T comparable] map[T]struct{}

// NewSet returns a new set with optional initial capacity.
func NewSet[T comparable](cap int) Set[T] {
	return make(Set[T], cap)
}

// Add adds v to the set.
func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

// Contains reports whether v is in the set.
func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

func GenerateGUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

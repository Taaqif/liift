// Package utils provides utility functions for internal use.
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
)

// GetEnv returns the value of the environment variable or the fallback value.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// MustGetEnv returns the value of the environment variable or fatals if unset.
func MustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

// GetEnvAsInt returns an environment variable parsed as int or the fallback.
func GetEnvAsInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

// Map applies fn to every element of in and returns the results.
func Map[T any, R any](in []T, fn func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

// Set is a set of comparable values backed by map[T]struct{}.
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

// GenerateGUID returns a random 32-character hex string.
func GenerateGUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

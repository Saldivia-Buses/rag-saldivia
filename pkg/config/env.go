// Package config provides centralized configuration for all SDA services.
//
// Two layers:
//   - Env/MustEnv: infrastructure config from environment variables (DB URLs, ports)
//   - Resolver: business config from Platform DB with scope cascade (tenant > plan > global)
package config

import (
	"fmt"
	"net/url"
	"os"
)

// Env reads an environment variable with a fallback default.
// Replaces the copy-pasted env() helper in every cmd/main.go.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// MustEnv reads an environment variable or panics if not set.
// Use for required infrastructure config (DB URLs, JWT keys).
func MustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}

// RedactURL strips credentials from a URL for safe logging.
// Returns the original string if parsing fails.
func RedactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}

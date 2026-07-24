// Package config loads runtime settings from the environment.
package config

import "os"

const defaultPort = "8080"

// Config holds process-wide server settings.
type Config struct {
	// Port is the TCP port the HTTP server listens on (without a host).
	Port string
	// DatabaseURL is an optional PostgreSQL connection string (DATABASE_URL).
	// Empty means the process should use non-database storage (e.g. in-memory).
	DatabaseURL string
}

// Load reads configuration from environment variables.
// PORT defaults to 8080 when unset or empty.
// DATABASE_URL is optional and may be empty.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

// Addr returns the listen address in host:port form.
func (c Config) Addr() string {
	return ":" + c.Port
}

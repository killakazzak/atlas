// Package config loads runtime settings from the environment.
package config

import (
	"os"
	"time"
)

const (
	defaultPort             = "8080"
	defaultJWTIssuer        = "atlas"
	defaultAccessTokenTTL   = 60 * time.Minute
)

// Config holds process-wide server settings.
type Config struct {
	// Port is the TCP port the HTTP server listens on (without a host).
	Port string
	// DatabaseURL is an optional PostgreSQL connection string (DATABASE_URL).
	// Empty means the process should use non-database storage (e.g. in-memory).
	DatabaseURL string

	// JWTSecret is the HMAC-SHA256 signing key for access tokens.
	JWTSecret string
	// JWTIssuer is placed in the "iss" claim.
	JWTIssuer string
	// JWTAccessTokenTTL controls how long an access token remains valid.
	JWTAccessTokenTTL time.Duration
}

// Load reads configuration from environment variables.
// PORT defaults to 8080 when unset or empty.
// DATABASE_URL is optional and may be empty.
// JWT_SECRET defaults to an insecure placeholder that warns at startup.
// JWT_ISSUER defaults to "atlas".
// JWT_ACCESS_TOKEN_TTL accepts Go duration strings (e.g. "1h"); defaults to 1h.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = defaultJWTIssuer
	}

	ttl := defaultAccessTokenTTL
	if raw := os.Getenv("JWT_ACCESS_TOKEN_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			ttl = d
		}
	}

	return Config{
		Port:              port,
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         jwtSecret,
		JWTIssuer:         jwtIssuer,
		JWTAccessTokenTTL: ttl,
	}
}

// Addr returns the listen address in host:port form.
func (c Config) Addr() string {
	return ":" + c.Port
}

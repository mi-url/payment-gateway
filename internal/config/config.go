// Package config loads and validates environment variables at startup.
// All configuration is read from environment variables — no config files.
// Missing required variables cause a fatal error at boot, not at runtime.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration, validated at startup.
type Config struct {
	// Server settings
	Port int
	Env  string // "development" or "production"

	// Database connection
	DatabaseURL string

	// GCP Cloud KMS
	KMSKeyResourceName string // e.g., "projects/P/locations/L/keyRings/R/cryptoKeys/K"

	// BNC-specific settings (QA vs Production)
	BNCBaseURL string

	// BNC webhook authentication — shared secret sent in x-api-key header.
	// Agreed upon during BNC onboarding/certification process.
	BNCWebhookAPIKey string

	// CORS — allowed origin for dashboard requests.
	CORSAllowedOrigin string
}

// Load reads configuration from environment variables and validates them.
// Returns an error describing all missing required variables at once.
func Load() (*Config, error) {
	var missing []string

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	kmsKey := os.Getenv("KMS_KEY_RESOURCE_NAME")
	if kmsKey == "" {
		missing = append(missing, "KMS_KEY_RESOURCE_NAME")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("PORT must be a valid integer: %w", err)
		}
		port = parsed
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	bncBaseURL := os.Getenv("BNC_BASE_URL")
	if bncBaseURL == "" {
		// Default to BNC QA environment
		bncBaseURL = "https://servicios.bncenlinea.com:16500/api"
	}

	bncWebhookKey := os.Getenv("BNC_WEBHOOK_API_KEY")

	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3001"
	}

	return &Config{
		Port:                port,
		Env:                 env,
		DatabaseURL:         dbURL,
		KMSKeyResourceName:  kmsKey,
		BNCBaseURL:          bncBaseURL,
		BNCWebhookAPIKey:    bncWebhookKey,
		CORSAllowedOrigin:   corsOrigin,
	}, nil
}

// IsProduction returns true if the application is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

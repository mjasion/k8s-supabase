package config

import (
	"time"
)

// ExtractorConfig holds configuration for the secret extractor
type ExtractorConfig struct {
	// Kubernetes configuration
	Namespace     string
	ServicePrefix string
	SecretName    string

	// Postgres configuration
	PostgresHost string
	PostgresPort int
	PostgresUser string

	// Retry configuration
	Retries       int
	RetryInterval time.Duration
	ReadyTimeout  time.Duration

	// URL configuration (from user-provided values)
	PublicURL string
	SiteURL   string

	// Kong configuration
	KongConfigPath string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *ExtractorConfig {
	return &ExtractorConfig{
		Namespace:      "default",
		ServicePrefix:  "supabase",
		SecretName:     "supabase-extracted-secrets",
		PostgresHost:   "postgres",
		PostgresPort:   5432,
		PostgresUser:   "postgres",
		Retries:        10,
		RetryInterval:  10 * time.Second,
		ReadyTimeout:   5 * time.Minute,
		KongConfigPath: "/home/kong/kong.yml",
	}
}

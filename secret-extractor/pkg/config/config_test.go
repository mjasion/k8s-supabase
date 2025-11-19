package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test default values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Namespace", cfg.Namespace, "default"},
		{"ServicePrefix", cfg.ServicePrefix, "supabase"},
		{"SecretName", cfg.SecretName, "supabase-extracted-secrets"},
		{"PostgresHost", cfg.PostgresHost, "postgres"},
		{"PostgresPort", cfg.PostgresPort, 5432},
		{"PostgresUser", cfg.PostgresUser, "postgres"},
		{"Retries", cfg.Retries, 10},
		{"RetryInterval", cfg.RetryInterval, 10 * time.Second},
		{"ReadyTimeout", cfg.ReadyTimeout, 5 * time.Minute},
		{"KongConfigPath", cfg.KongConfigPath, "/home/kong/kong.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestExtractorConfigCustomization(t *testing.T) {
	cfg := &ExtractorConfig{
		Namespace:      "custom-namespace",
		ServicePrefix:  "my-supabase",
		SecretName:     "my-secrets",
		PostgresHost:   "custom-postgres",
		PostgresPort:   5433,
		PostgresUser:   "custom-user",
		Retries:        20,
		RetryInterval:  5 * time.Second,
		ReadyTimeout:   10 * time.Minute,
		PublicURL:      "https://api.example.com",
		SiteURL:        "https://app.example.com",
		KongConfigPath: "/custom/path/kong.yml",
	}

	if cfg.Namespace != "custom-namespace" {
		t.Errorf("Namespace: got %v, want custom-namespace", cfg.Namespace)
	}
	if cfg.ServicePrefix != "my-supabase" {
		t.Errorf("ServicePrefix: got %v, want my-supabase", cfg.ServicePrefix)
	}
	if cfg.PostgresPort != 5433 {
		t.Errorf("PostgresPort: got %v, want 5433", cfg.PostgresPort)
	}
	if cfg.Retries != 20 {
		t.Errorf("Retries: got %v, want 20", cfg.Retries)
	}
	if cfg.PublicURL != "https://api.example.com" {
		t.Errorf("PublicURL: got %v, want https://api.example.com", cfg.PublicURL)
	}
}

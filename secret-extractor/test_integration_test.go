// +build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/config"
	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/extractor"
)

// Integration tests require a running Kubernetes cluster with Supabase deployed
// Run with: go test -tags=integration

func TestIntegrationExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.ExtractorConfig{
		Namespace:      "supabase-test",
		ServicePrefix:  "supabase",
		SecretName:     "test-extracted-secrets",
		PostgresHost:   "supabase-db",
		PostgresPort:   5432,
		PostgresUser:   "postgres",
		Retries:        3,
		RetryInterval:  5 * time.Second,
		ReadyTimeout:   2 * time.Minute,
		PublicURL:      "http://localhost:8000",
		SiteURL:        "http://localhost:3000",
		KongConfigPath: "/home/kong/kong.yml",
	}

	e, err := extractor.NewExtractor(cfg)
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	ctx := context.Background()
	secrets, err := e.Extract(ctx)
	if err != nil {
		t.Fatalf("Failed to extract secrets: %v", err)
	}

	// Verify required secrets
	if secrets.PostgresPassword == "" {
		t.Error("PostgresPassword is empty")
	}
	if secrets.JWTSecret == "" {
		t.Error("JWTSecret is empty")
	}
	if secrets.AnonKey == "" {
		t.Error("AnonKey is empty")
	}
	if secrets.ServiceRoleKey == "" {
		t.Error("ServiceRoleKey is empty")
	}

	// Test persistence
	err = e.PersistSecrets(ctx, cfg.SecretName, secrets)
	if err != nil {
		t.Fatalf("Failed to persist secrets: %v", err)
	}

	t.Logf("Successfully extracted and persisted secrets to %s", cfg.SecretName)
}

func TestIntegrationValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()
	cfg.Namespace = "supabase-test"
	cfg.PublicURL = "http://localhost:8000"
	cfg.SiteURL = "http://localhost:3000"

	e, err := extractor.NewExtractor(cfg)
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	ctx := context.Background()
	secrets, err := e.Extract(ctx)
	if err != nil {
		t.Fatalf("Failed to extract secrets: %v", err)
	}

	// Test JWT validation
	err = e.ValidateJWTTokens(secrets)
	if err != nil {
		t.Errorf("JWT validation failed: %v", err)
	}

	// Test Postgres connection validation
	err = e.ValidatePostgresConnection(secrets)
	if err != nil {
		t.Errorf("Postgres connection validation failed: %v", err)
	}
}

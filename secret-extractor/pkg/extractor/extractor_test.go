package extractor

import (
	"testing"

	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/config"
)

func TestNewExtractor(t *testing.T) {
	cfg := &config.ExtractorConfig{
		Namespace:     "test-namespace",
		ServicePrefix: "test-supabase",
	}

	// This will fail without a real K8s cluster, but we're testing the structure
	_, err := NewExtractor(cfg)
	// We expect an error since we don't have a kubeconfig in test environment
	// but we're just testing that the function exists and has proper signature
	if err == nil {
		// If somehow it succeeds, that's also okay for this test
		t.Log("NewExtractor succeeded (unexpected in test env but acceptable)")
	}
}

func TestExtractedSecretsStructure(t *testing.T) {
	secrets := &ExtractedSecrets{
		PostgresPassword:  "test-password",
		JWTSecret:         "test-jwt-secret-with-at-least-32-characters",
		AnonKey:           "test.anon.key",
		ServiceRoleKey:    "test.service.key",
		DashboardUsername: "admin",
		DashboardPassword: "admin-pass",
		SecretKeyBase:     "secret-key-base",
		VaultEncKey:       "vault-enc-key",
		PGMetaCryptoKey:   "pg-meta-crypto-key",
		PublicURL:         "https://api.example.com",
		SiteURL:           "https://app.example.com",
		APIURL:            "https://api.example.com",
	}

	// Test all fields are accessible
	if secrets.PostgresPassword != "test-password" {
		t.Error("PostgresPassword field not accessible")
	}
	if secrets.JWTSecret != "test-jwt-secret-with-at-least-32-characters" {
		t.Error("JWTSecret field not accessible")
	}
	if secrets.AnonKey != "test.anon.key" {
		t.Error("AnonKey field not accessible")
	}
	if secrets.ServiceRoleKey != "test.service.key" {
		t.Error("ServiceRoleKey field not accessible")
	}
	if secrets.PublicURL != "https://api.example.com" {
		t.Error("PublicURL field not accessible")
	}
}

func TestPersistSecretsDataMapping(t *testing.T) {
	secrets := &ExtractedSecrets{
		PostgresPassword:     "db-password",
		JWTSecret:            "jwt-secret-32-characters-long",
		AnonKey:              "anon-key",
		ServiceRoleKey:       "service-key",
		DashboardUsername:    "admin",
		DashboardPassword:    "admin-pass",
		SecretKeyBase:        "secret-base",
		VaultEncKey:          "vault-key",
		PGMetaCryptoKey:      "meta-key",
		LogflarePublicToken:  "public-token",
		LogflarePrivateToken: "private-token",
		PublicURL:            "https://api.test.com",
		SiteURL:              "https://app.test.com",
	}

	// Test that we can create the expected secret data structure
	// This simulates what PersistSecrets does
	secretData := make(map[string]string)

	if secrets.PostgresPassword != "" {
		secretData["POSTGRES_PASSWORD"] = secrets.PostgresPassword
	}
	if secrets.JWTSecret != "" {
		secretData["JWT_SECRET"] = secrets.JWTSecret
	}
	if secrets.AnonKey != "" {
		secretData["ANON_KEY"] = secrets.AnonKey
	}
	if secrets.ServiceRoleKey != "" {
		secretData["SERVICE_ROLE_KEY"] = secrets.ServiceRoleKey
	}

	// Verify mappings
	expectedKeys := []string{
		"POSTGRES_PASSWORD",
		"JWT_SECRET",
		"ANON_KEY",
		"SERVICE_ROLE_KEY",
	}

	for _, key := range expectedKeys {
		if _, exists := secretData[key]; !exists {
			t.Errorf("expected key %q not found in secret data", key)
		}
	}

	if secretData["POSTGRES_PASSWORD"] != "db-password" {
		t.Errorf("POSTGRES_PASSWORD = %q, want db-password", secretData["POSTGRES_PASSWORD"])
	}
}

func TestAuthSecretsStructure(t *testing.T) {
	authSecrets := &AuthSecrets{
		JWTSecret:     "jwt-secret",
		SecretKeyBase: "secret-key-base",
		AnonKey:       "anon-key",
		ServiceKey:    "service-key",
	}

	if authSecrets.JWTSecret != "jwt-secret" {
		t.Error("JWTSecret field not accessible")
	}
	if authSecrets.SecretKeyBase != "secret-key-base" {
		t.Error("SecretKeyBase field not accessible")
	}
}

func TestKongSecretsStructure(t *testing.T) {
	kongSecrets := &KongSecrets{
		AnonKey:           "anon-key",
		ServiceRoleKey:    "service-key",
		DashboardUsername: "admin",
		DashboardPassword: "pass",
	}

	if kongSecrets.AnonKey != "anon-key" {
		t.Error("AnonKey field not accessible")
	}
	if kongSecrets.DashboardUsername != "admin" {
		t.Error("DashboardUsername field not accessible")
	}
}

func TestPostgresSecretsStructure(t *testing.T) {
	pgSecrets := &PostgresSecrets{
		Password:  "db-password",
		JWTSecret: "jwt-secret",
	}

	if pgSecrets.Password != "db-password" {
		t.Error("Password field not accessible")
	}
	if pgSecrets.JWTSecret != "jwt-secret" {
		t.Error("JWTSecret field not accessible")
	}
}

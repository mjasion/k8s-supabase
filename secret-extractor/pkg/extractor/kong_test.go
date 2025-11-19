package extractor

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseKongConfig(t *testing.T) {
	kongYAML := `
_format_version: "3.0"
consumers:
  - username: anon
    keyauth_credentials:
      - key: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.anon.key
  - username: service_role
    keyauth_credentials:
      - key: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.service.key
  - username: dashboard
    basicauth_credentials:
      - username: admin
        password: admin-password
`

	e := &Extractor{}
	secrets, err := e.parseKongConfig([]byte(kongYAML))
	if err != nil {
		t.Fatalf("parseKongConfig failed: %v", err)
	}

	if secrets.AnonKey != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.anon.key" {
		t.Errorf("AnonKey = %q, want %q", secrets.AnonKey, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.anon.key")
	}

	if secrets.ServiceRoleKey != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.service.key" {
		t.Errorf("ServiceRoleKey = %q, want %q", secrets.ServiceRoleKey, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.service.key")
	}

	if secrets.DashboardUsername != "admin" {
		t.Errorf("DashboardUsername = %q, want admin", secrets.DashboardUsername)
	}

	if secrets.DashboardPassword != "admin-password" {
		t.Errorf("DashboardPassword = %q, want admin-password", secrets.DashboardPassword)
	}
}

func TestParseKongConfigEmpty(t *testing.T) {
	kongYAML := `
_format_version: "3.0"
consumers: []
`

	e := &Extractor{}
	secrets, err := e.parseKongConfig([]byte(kongYAML))
	if err != nil {
		t.Fatalf("parseKongConfig failed: %v", err)
	}

	if secrets.AnonKey != "" {
		t.Errorf("expected empty AnonKey, got %q", secrets.AnonKey)
	}

	if secrets.ServiceRoleKey != "" {
		t.Errorf("expected empty ServiceRoleKey, got %q", secrets.ServiceRoleKey)
	}
}

func TestParseKongConfigInvalid(t *testing.T) {
	kongYAML := `invalid yaml: [[[`

	e := &Extractor{}
	_, err := e.parseKongConfig([]byte(kongYAML))
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestKongConfigStructure(t *testing.T) {
	// Test the KongConfig struct can be properly unmarshaled
	kongYAML := `
_format_version: "3.0"
consumers:
  - username: test-user
    keyauth_credentials:
      - key: test-key-123
services:
  - name: test-service
    url: http://backend:8080
    routes:
      - name: test-route
        paths:
          - /api
    plugins:
      - name: cors
        config:
          origins:
            - "*"
`

	var config KongConfig
	err := yaml.Unmarshal([]byte(kongYAML), &config)
	if err != nil {
		t.Fatalf("failed to unmarshal Kong config: %v", err)
	}

	if config.FormatVersion != "3.0" {
		t.Errorf("FormatVersion = %q, want 3.0", config.FormatVersion)
	}

	if len(config.Consumers) != 1 {
		t.Errorf("expected 1 consumer, got %d", len(config.Consumers))
	}

	if config.Consumers[0].Username != "test-user" {
		t.Errorf("consumer username = %q, want test-user", config.Consumers[0].Username)
	}

	if len(config.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(config.Services))
	}
}

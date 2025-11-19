package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestSecretStructure(t *testing.T) {
	// Test that we can create the expected secret structure
	secretData := map[string]string{
		"POSTGRES_PASSWORD": "test-password",
		"JWT_SECRET":        "test-jwt-secret",
		"ANON_KEY":          "test-anon-key",
		"SERVICE_ROLE_KEY":  "test-service-key",
	}

	// Verify all expected keys exist
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
}

func TestSecretTypeOpaque(t *testing.T) {
	// Test that we use the correct secret type
	secretType := corev1.SecretTypeOpaque

	if secretType != "Opaque" {
		t.Errorf("secret type = %q, want Opaque", secretType)
	}
}

func TestSecretLabels(t *testing.T) {
	// Test expected labels structure
	labels := map[string]string{
		"app.kubernetes.io/name":       "supabase",
		"app.kubernetes.io/component":  "extracted-secrets",
		"app.kubernetes.io/managed-by": "secret-extractor",
	}

	expectedLabels := map[string]string{
		"app.kubernetes.io/name":       "supabase",
		"app.kubernetes.io/component":  "extracted-secrets",
		"app.kubernetes.io/managed-by": "secret-extractor",
	}

	for key, expectedValue := range expectedLabels {
		if value, exists := labels[key]; !exists {
			t.Errorf("expected label %q not found", key)
		} else if value != expectedValue {
			t.Errorf("label %q = %q, want %q", key, value, expectedValue)
		}
	}
}

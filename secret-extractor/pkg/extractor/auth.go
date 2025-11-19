package extractor

import (
	"context"
	"fmt"
	"strings"
)

// AuthSecrets holds secrets extracted from the Auth service
type AuthSecrets struct {
	JWTSecret     string
	SecretKeyBase string
	AnonKey       string
	ServiceKey    string
}

// extractFromAuth extracts secrets from the Auth (GoTrue) service
func (e *Extractor) extractFromAuth(ctx context.Context) (*AuthSecrets, error) {
	// Find the auth pod
	podName, containerName, err := e.k8s.FindFirstReadyPod(ctx,
		fmt.Sprintf("app.kubernetes.io/component=%s-auth", e.config.ServicePrefix))
	if err != nil {
		// Try alternative label format
		podName, containerName, err = e.k8s.FindFirstReadyPod(ctx,
			fmt.Sprintf("app=%s-auth", e.config.ServicePrefix))
		if err != nil {
			return nil, fmt.Errorf("failed to find auth pod: %w", err)
		}
	}

	// Get all environment variables
	envVars, err := e.k8s.GetAllPodEnvVars(ctx, podName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth env vars: %w", err)
	}

	secrets := &AuthSecrets{}

	// Map environment variables to secrets
	for key, value := range envVars {
		keyUpper := strings.ToUpper(key)
		switch {
		case keyUpper == "GOTRUE_JWT_SECRET" || keyUpper == "JWT_SECRET":
			secrets.JWTSecret = value
		case keyUpper == "SECRET_KEY_BASE":
			secrets.SecretKeyBase = value
		case strings.Contains(keyUpper, "ANON_KEY"):
			secrets.AnonKey = value
		case strings.Contains(keyUpper, "SERVICE_ROLE_KEY") || strings.Contains(keyUpper, "SERVICE_KEY"):
			secrets.ServiceKey = value
		}
	}

	return secrets, nil
}

// extractFromMeta extracts secrets from the Meta service
func (e *Extractor) extractFromMeta(ctx context.Context) (map[string]string, error) {
	// Find the meta pod
	podName, containerName, err := e.k8s.FindFirstReadyPod(ctx,
		fmt.Sprintf("app.kubernetes.io/component=%s-meta", e.config.ServicePrefix))
	if err != nil {
		// Try alternative label format
		podName, containerName, err = e.k8s.FindFirstReadyPod(ctx,
			fmt.Sprintf("app=%s-meta", e.config.ServicePrefix))
		if err != nil {
			return nil, fmt.Errorf("failed to find meta pod: %w", err)
		}
	}

	// Get specific environment variables
	envVars, err := e.k8s.GetPodEnvVars(ctx, podName, containerName, []string{
		"PG_META_CRYPTO_KEY",
		"PG_META_DB_PASSWORD",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get meta env vars: %w", err)
	}

	return envVars, nil
}

// extractFromStorage extracts secrets from the Storage service
func (e *Extractor) extractFromStorage(ctx context.Context) (map[string]string, error) {
	// Find the storage pod
	podName, containerName, err := e.k8s.FindFirstReadyPod(ctx,
		fmt.Sprintf("app.kubernetes.io/component=%s-storage", e.config.ServicePrefix))
	if err != nil {
		// Try alternative label format
		podName, containerName, err = e.k8s.FindFirstReadyPod(ctx,
			fmt.Sprintf("app=%s-storage", e.config.ServicePrefix))
		if err != nil {
			// Storage might not be deployed, return empty
			return make(map[string]string), nil
		}
	}

	// Get specific environment variables
	envVars, err := e.k8s.GetPodEnvVars(ctx, podName, containerName, []string{
		"ANON_KEY",
		"SERVICE_KEY",
		"PGRST_JWT_SECRET",
		"DATABASE_URL",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get storage env vars: %w", err)
	}

	return envVars, nil
}

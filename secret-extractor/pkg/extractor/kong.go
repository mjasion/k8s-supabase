package extractor

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// KongSecrets holds secrets extracted from Kong
type KongSecrets struct {
	AnonKey           string
	ServiceRoleKey    string
	DashboardUsername string
	DashboardPassword string
}

// KongConfig represents the Kong declarative configuration
type KongConfig struct {
	FormatVersion string `yaml:"_format_version"`
	Consumers     []struct {
		Username    string `yaml:"username"`
		KeyAuth     []struct {
			Key string `yaml:"key"`
		} `yaml:"keyauth_credentials"`
		BasicAuth []struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"basicauth_credentials"`
	} `yaml:"consumers"`
	Services []struct {
		Name   string `yaml:"name"`
		URL    string `yaml:"url"`
		Routes []struct {
			Name  string   `yaml:"name"`
			Paths []string `yaml:"paths"`
		} `yaml:"routes"`
		Plugins []struct {
			Name   string                 `yaml:"name"`
			Config map[string]interface{} `yaml:"config"`
		} `yaml:"plugins"`
	} `yaml:"services"`
}

// extractFromKong extracts secrets from Kong configuration
func (e *Extractor) extractFromKong(ctx context.Context) (*KongSecrets, error) {
	// Find the kong pod
	podName, containerName, err := e.k8s.FindFirstReadyPod(ctx,
		fmt.Sprintf("app.kubernetes.io/component=%s-kong", e.config.ServicePrefix))
	if err != nil {
		// Try alternative label format
		podName, containerName, err = e.k8s.FindFirstReadyPod(ctx,
			fmt.Sprintf("app=%s-kong", e.config.ServicePrefix))
		if err != nil {
			return nil, fmt.Errorf("failed to find kong pod: %w", err)
		}
	}

	secrets := &KongSecrets{}

	// Try to read kong.yml configuration
	kongYAML, err := e.k8s.ReadFileFromPod(ctx, podName, containerName, e.config.KongConfigPath)
	if err == nil {
		// Parse Kong config
		if kongSecrets, err := e.parseKongConfig(kongYAML); err == nil {
			secrets = kongSecrets
		}
	}

	// If we didn't get secrets from config, try environment variables
	if secrets.AnonKey == "" || secrets.ServiceRoleKey == "" {
		envSecrets, err := e.extractKongFromEnv(ctx, podName, containerName)
		if err == nil {
			if secrets.AnonKey == "" {
				secrets.AnonKey = envSecrets.AnonKey
			}
			if secrets.ServiceRoleKey == "" {
				secrets.ServiceRoleKey = envSecrets.ServiceRoleKey
			}
			if secrets.DashboardUsername == "" {
				secrets.DashboardUsername = envSecrets.DashboardUsername
			}
			if secrets.DashboardPassword == "" {
				secrets.DashboardPassword = envSecrets.DashboardPassword
			}
		}
	}

	return secrets, nil
}

// parseKongConfig parses the Kong YAML configuration
func (e *Extractor) parseKongConfig(data []byte) (*KongSecrets, error) {
	var config KongConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse kong config: %w", err)
	}

	secrets := &KongSecrets{}

	// Extract API keys from consumers
	for _, consumer := range config.Consumers {
		username := strings.ToLower(consumer.Username)

		// Check KeyAuth credentials
		if len(consumer.KeyAuth) > 0 {
			key := consumer.KeyAuth[0].Key
			switch username {
			case "anon":
				secrets.AnonKey = key
			case "service_role", "service-role":
				secrets.ServiceRoleKey = key
			}
		}

		// Check BasicAuth credentials (for dashboard)
		if len(consumer.BasicAuth) > 0 {
			if strings.Contains(username, "dashboard") {
				secrets.DashboardUsername = consumer.BasicAuth[0].Username
				secrets.DashboardPassword = consumer.BasicAuth[0].Password
			}
		}
	}

	return secrets, nil
}

// extractKongFromEnv extracts Kong secrets from environment variables
func (e *Extractor) extractKongFromEnv(ctx context.Context, podName, containerName string) (*KongSecrets, error) {
	envVars, err := e.k8s.GetAllPodEnvVars(ctx, podName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get kong env vars: %w", err)
	}

	secrets := &KongSecrets{}

	// Map common environment variable names to secrets
	for key, value := range envVars {
		keyUpper := strings.ToUpper(key)
		switch {
		case strings.Contains(keyUpper, "ANON_KEY"):
			secrets.AnonKey = value
		case strings.Contains(keyUpper, "SERVICE_ROLE_KEY") || strings.Contains(keyUpper, "SERVICE_KEY"):
			secrets.ServiceRoleKey = value
		case strings.Contains(keyUpper, "DASHBOARD_USERNAME"):
			secrets.DashboardUsername = value
		case strings.Contains(keyUpper, "DASHBOARD_PASSWORD"):
			secrets.DashboardPassword = value
		}
	}

	return secrets, nil
}

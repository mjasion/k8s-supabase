package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// DockerCompose represents a docker-compose configuration
type DockerCompose struct {
	Version  string                    `yaml:"version"`
	Services map[string]*Service       `yaml:"services"`
	Networks map[string]*Network       `yaml:"networks"`
	Volumes  map[string]*Volume        `yaml:"volumes"`
}

// Service represents a docker-compose service
type Service struct {
	Image         string                 `yaml:"image"`
	Container     string                 `yaml:"container_name"`
	Command       interface{}            `yaml:"command"` // Can be string or []string
	Environment   map[string]interface{} `yaml:"environment"`
	EnvFile       interface{}            `yaml:"env_file"` // Can be string or []string
	Ports         []string               `yaml:"ports"`
	Volumes       []string               `yaml:"volumes"`
	DependsOn     interface{}            `yaml:"depends_on"` // Can be []string or map
	Restart       string                 `yaml:"restart"`
	HealthCheck   *HealthCheck           `yaml:"healthcheck"`
	Networks      interface{}            `yaml:"networks"` // Can be []string or map
	Labels        map[string]string      `yaml:"labels"`
	Entrypoint    interface{}            `yaml:"entrypoint"` // Can be string or []string
	WorkingDir    string                 `yaml:"working_dir"`
	User          string                 `yaml:"user"`
	CapAdd        []string               `yaml:"cap_add"`
	CapDrop       []string               `yaml:"cap_drop"`
	Privileged    bool                   `yaml:"privileged"`
	SecurityOpt   []string               `yaml:"security_opt"`
	Tmpfs         interface{}            `yaml:"tmpfs"` // Can be []string or map
}

// HealthCheck represents a docker-compose health check
type HealthCheck struct {
	Test     interface{} `yaml:"test"` // Can be string or []string
	Interval string      `yaml:"interval"`
	Timeout  string      `yaml:"timeout"`
	Retries  int         `yaml:"retries"`
}

// Network represents a docker-compose network
type Network struct {
	Driver     string            `yaml:"driver"`
	External   bool              `yaml:"external"`
	DriverOpts map[string]string `yaml:"driver_opts"`
}

// Volume represents a docker-compose volume
type Volume struct {
	Driver     string            `yaml:"driver"`
	External   bool              `yaml:"external"`
	DriverOpts map[string]string `yaml:"driver_opts"`
}

// EnvVar represents an environment variable with metadata
type EnvVar struct {
	Name         string
	DefaultValue string
	Description  string
	IsSecret     bool
}

// Parser parses docker-compose files
type Parser struct{}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses docker-compose.yml and .env.example
func (p *Parser) Parse(composeData []byte, envData []byte) (*DockerCompose, error) {
	var config DockerCompose
	if err := yaml.Unmarshal(composeData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	// Parse environment variables if provided
	if len(envData) > 0 {
		envVars := p.parseEnvFile(envData)
		p.enrichServicesWithEnvVars(&config, envVars)
	}

	return &config, nil
}

// parseEnvFile parses .env.example file
func (p *Parser) parseEnvFile(data []byte) map[string]*EnvVar {
	envVars := make(map[string]*EnvVar)
	lines := strings.Split(string(data), "\n")

	var currentComment string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			currentComment = ""
			continue
		}

		// Collect comments
		if strings.HasPrefix(line, "#") {
			if currentComment != "" {
				currentComment += " "
			}
			currentComment += strings.TrimPrefix(line, "#")
			continue
		}

		// Parse environment variable
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Detect if it's a secret (password, key, token, secret)
			isSecret := containsSecretKeyword(name)

			envVars[name] = &EnvVar{
				Name:         name,
				DefaultValue: value,
				Description:  strings.TrimSpace(currentComment),
				IsSecret:     isSecret,
			}

			currentComment = ""
		}
	}

	return envVars
}

// enrichServicesWithEnvVars enriches service environment with parsed env vars
func (p *Parser) enrichServicesWithEnvVars(config *DockerCompose, envVars map[string]*EnvVar) {
	// This method can be used to add metadata to services based on env vars
	// For now, we'll keep it simple and let the converter handle env vars
}

// GetDependencies returns service dependencies
func (s *Service) GetDependencies() []string {
	if s.DependsOn == nil {
		return []string{}
	}

	switch v := s.DependsOn.(type) {
	case []interface{}:
		deps := make([]string, len(v))
		for i, dep := range v {
			deps[i] = fmt.Sprintf("%v", dep)
		}
		return deps
	case map[string]interface{}:
		deps := make([]string, 0, len(v))
		for dep := range v {
			deps = append(deps, dep)
		}
		return deps
	case []string:
		return v
	default:
		return []string{}
	}
}

// GetCommandAsSlice returns command as a string slice
func (s *Service) GetCommandAsSlice() []string {
	if s.Command == nil {
		return []string{}
	}

	switch v := s.Command.(type) {
	case []interface{}:
		cmd := make([]string, len(v))
		for i, c := range v {
			cmd[i] = fmt.Sprintf("%v", c)
		}
		return cmd
	case []string:
		return v
	case string:
		return []string{"/bin/sh", "-c", v}
	default:
		return []string{}
	}
}

// GetEntrypointAsSlice returns entrypoint as a string slice
func (s *Service) GetEntrypointAsSlice() []string {
	if s.Entrypoint == nil {
		return []string{}
	}

	switch v := s.Entrypoint.(type) {
	case []interface{}:
		ep := make([]string, len(v))
		for i, e := range v {
			ep[i] = fmt.Sprintf("%v", e)
		}
		return ep
	case []string:
		return v
	case string:
		return []string{v}
	default:
		return []string{}
	}
}

// containsSecretKeyword checks if a string contains secret-related keywords
func containsSecretKeyword(s string) bool {
	s = strings.ToLower(s)
	secretKeywords := []string{"password", "secret", "key", "token", "credential", "api_key"}
	for _, keyword := range secretKeywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

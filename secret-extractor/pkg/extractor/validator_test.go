package extractor

import (
	"testing"
)

func TestIsJWT(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid JWT format",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expected: true,
		},
		{
			name:     "invalid JWT - missing parts",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			expected: false,
		},
		{
			name:     "invalid JWT - too many parts",
			token:    "a.b.c.d",
			expected: false,
		},
		{
			name:     "empty string",
			token:    "",
			expected: false,
		},
		{
			name:     "plain string",
			token:    "not-a-jwt-token",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJWT(tt.token)
			if result != tt.expected {
				t.Errorf("isJWT(%q) = %v, want %v", tt.token, result, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Mock extractor
	e := &Extractor{}

	tests := []struct {
		name        string
		secrets     *ExtractedSecrets
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid secrets",
			secrets: &ExtractedSecrets{
				PostgresPassword:  "super-secret-password",
				JWTSecret:         "super-secret-jwt-token-with-at-least-32-characters-long",
				AnonKey:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI",
				ServiceRoleKey:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
				DashboardUsername: "admin",
				DashboardPassword: "admin-password",
				PublicURL:         "https://api.example.com",
				SiteURL:           "https://app.example.com",
			},
			expectError: false,
		},
		{
			name: "missing postgres password",
			secrets: &ExtractedSecrets{
				PostgresPassword: "",
				JWTSecret:        "super-secret-jwt-token-with-at-least-32-characters-long",
				AnonKey:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI",
				ServiceRoleKey:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
				PublicURL:        "https://api.example.com",
				SiteURL:          "https://app.example.com",
			},
			expectError: true,
			errorMsg:    "POSTGRES_PASSWORD is empty",
		},
		{
			name: "JWT secret too short",
			secrets: &ExtractedSecrets{
				PostgresPassword: "super-secret-password",
				JWTSecret:        "short",
				AnonKey:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI",
				ServiceRoleKey:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
				PublicURL:        "https://api.example.com",
				SiteURL:          "https://app.example.com",
			},
			expectError: true,
			errorMsg:    "JWT_SECRET is too short",
		},
		{
			name: "invalid anon key format",
			secrets: &ExtractedSecrets{
				PostgresPassword: "super-secret-password",
				JWTSecret:        "super-secret-jwt-token-with-at-least-32-characters-long",
				AnonKey:          "not-a-jwt",
				ServiceRoleKey:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
				PublicURL:        "https://api.example.com",
				SiteURL:          "https://app.example.com",
			},
			expectError: true,
			errorMsg:    "ANON_KEY does not appear to be a valid JWT",
		},
		{
			name: "missing public URL",
			secrets: &ExtractedSecrets{
				PostgresPassword: "super-secret-password",
				JWTSecret:        "super-secret-jwt-token-with-at-least-32-characters-long",
				AnonKey:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI",
				ServiceRoleKey:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
				PublicURL:        "",
				SiteURL:          "https://app.example.com",
			},
			expectError: true,
			errorMsg:    "PUBLIC_URL is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := e.validate(tt.secrets)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateJWTTokens(t *testing.T) {
	e := &Extractor{}

	tests := []struct {
		name        string
		secrets     *ExtractedSecrets
		expectError bool
	}{
		{
			name: "valid JWT tokens",
			secrets: &ExtractedSecrets{
				AnonKey:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI",
				ServiceRoleKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
			},
			expectError: false,
		},
		{
			name: "invalid anon key",
			secrets: &ExtractedSecrets{
				AnonKey:        "not-a-jwt",
				ServiceRoleKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.M2d2z4SFn5C7HlJlaSLfrzuYim9nbY_XI40uWFN3hEE",
			},
			expectError: true,
		},
		{
			name: "invalid service role key",
			secrets: &ExtractedSecrets{
				AnonKey:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI",
				ServiceRoleKey: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := e.ValidateJWTTokens(tt.secrets)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePostgresConnection(t *testing.T) {
	e := &Extractor{}

	tests := []struct {
		name        string
		secrets     *ExtractedSecrets
		expectError bool
	}{
		{
			name: "valid password",
			secrets: &ExtractedSecrets{
				PostgresPassword: "secure-password",
			},
			expectError: false,
		},
		{
			name: "empty password",
			secrets: &ExtractedSecrets{
				PostgresPassword: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := e.ValidatePostgresConnection(tt.secrets)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

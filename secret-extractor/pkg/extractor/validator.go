package extractor

import (
	"fmt"
	"strings"
)

// validate validates that all required secrets are present and valid
func (e *Extractor) validate(secrets *ExtractedSecrets) error {
	var errors []string

	// Required secrets
	if secrets.PostgresPassword == "" {
		errors = append(errors, "POSTGRES_PASSWORD is empty")
	}

	if secrets.AnonKey == "" {
		errors = append(errors, "ANON_KEY is empty")
	}

	if secrets.ServiceRoleKey == "" {
		errors = append(errors, "SERVICE_ROLE_KEY is empty")
	}

	// JWT secret is critical
	if secrets.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is empty")
	}

	// Validate JWT secret length (should be at least 32 characters)
	if len(secrets.JWTSecret) < 32 {
		errors = append(errors, fmt.Sprintf("JWT_SECRET is too short (%d chars, minimum 32 required)", len(secrets.JWTSecret)))
	}

	// Validate Anon Key format (should be a JWT)
	if secrets.AnonKey != "" && !isJWT(secrets.AnonKey) {
		errors = append(errors, "ANON_KEY does not appear to be a valid JWT")
	}

	// Validate Service Role Key format (should be a JWT)
	if secrets.ServiceRoleKey != "" && !isJWT(secrets.ServiceRoleKey) {
		errors = append(errors, "SERVICE_ROLE_KEY does not appear to be a valid JWT")
	}

	// Dashboard credentials (optional but recommended)
	if secrets.DashboardUsername == "" || secrets.DashboardPassword == "" {
		errors = append(errors, "Warning: Dashboard credentials not found (DASHBOARD_USERNAME or DASHBOARD_PASSWORD)")
	}

	// URLs are required
	if secrets.PublicURL == "" {
		errors = append(errors, "PUBLIC_URL is empty")
	}

	if secrets.SiteURL == "" {
		errors = append(errors, "SITE_URL is empty")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// isJWT checks if a string looks like a JWT token
func isJWT(token string) bool {
	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	return len(parts) == 3
}

// ValidatePostgresConnection validates that we can connect to Postgres with extracted credentials
func (e *Extractor) ValidatePostgresConnection(secrets *ExtractedSecrets) error {
	// This would be implemented to actually test the connection
	// For now, we just validate the password is not empty
	if secrets.PostgresPassword == "" {
		return fmt.Errorf("postgres password is empty")
	}
	return nil
}

// ValidateJWTTokens validates that the JWT tokens are correctly derived from the secret
func (e *Extractor) ValidateJWTTokens(secrets *ExtractedSecrets) error {
	// This would implement actual JWT validation
	// For now, we just check they exist and are JWT-formatted
	if !isJWT(secrets.AnonKey) {
		return fmt.Errorf("anon key is not a valid JWT")
	}
	if !isJWT(secrets.ServiceRoleKey) {
		return fmt.Errorf("service role key is not a valid JWT")
	}
	return nil
}

package extractor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// PostgresSecrets holds secrets extracted from Postgres
type PostgresSecrets struct {
	Password  string
	JWTSecret string
}

// extractFromPostgres extracts secrets from the Postgres database
func (e *Extractor) extractFromPostgres(ctx context.Context) (*PostgresSecrets, error) {
	// Find the postgres pod
	podName, containerName, err := e.k8s.FindFirstReadyPod(ctx,
		fmt.Sprintf("app.kubernetes.io/component=%s-db", e.config.ServicePrefix))
	if err != nil {
		// Try alternative label format
		podName, containerName, err = e.k8s.FindFirstReadyPod(ctx,
			fmt.Sprintf("app=%s-db", e.config.ServicePrefix))
		if err != nil {
			return nil, fmt.Errorf("failed to find postgres pod: %w", err)
		}
	}

	// Get postgres password from environment
	password, err := e.k8s.GetPodEnvVar(ctx, podName, containerName, "POSTGRES_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres password: %w", err)
	}

	if password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is empty")
	}

	// Connect to Postgres
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable connect_timeout=10",
		e.config.PostgresHost,
		e.config.PostgresPort,
		e.config.PostgresUser,
		password,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	secrets := &PostgresSecrets{
		Password: password,
	}

	// Try to extract JWT secret from database settings
	// This may not exist in all setups, so we don't fail if it's not found
	var jwtSecret sql.NullString
	err = db.QueryRowContext(ctx,
		"SELECT current_setting('app.settings.jwt_secret', true)").Scan(&jwtSecret)
	if err == nil && jwtSecret.Valid {
		secrets.JWTSecret = jwtSecret.String
	}

	// Alternatively, try to get JWT secret from pgrst schema settings
	err = db.QueryRowContext(ctx, `
		SELECT value FROM pgrst.config WHERE name = 'jwt_secret'
	`).Scan(&jwtSecret)
	if err == nil && jwtSecret.Valid {
		secrets.JWTSecret = jwtSecret.String
	}

	return secrets, nil
}

// getPostgresPasswordFromPod retrieves the postgres password from the pod environment
func (e *Extractor) getPostgresPasswordFromPod(ctx context.Context) (string, error) {
	// Find the postgres pod
	podName := fmt.Sprintf("%s-db-0", e.config.ServicePrefix)

	// Get the password from environment
	password, err := e.k8s.GetPodEnvVar(ctx, podName, "postgres", "POSTGRES_PASSWORD")
	if err != nil {
		// Try with just the service prefix
		podName = fmt.Sprintf("%s-postgres-0", e.config.ServicePrefix)
		password, err = e.k8s.GetPodEnvVar(ctx, podName, "postgres", "POSTGRES_PASSWORD")
		if err != nil {
			return "", fmt.Errorf("failed to get postgres password from pod: %w", err)
		}
	}

	return strings.TrimSpace(password), nil
}

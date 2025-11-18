package extractor

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/config"
	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/k8s"
)

// ExtractedSecrets holds all extracted secrets from Supabase services
type ExtractedSecrets struct {
	// Core secrets
	PostgresPassword string
	JWTSecret        string
	AnonKey          string
	ServiceRoleKey   string

	// Dashboard
	DashboardUsername string
	DashboardPassword string

	// Encryption keys
	SecretKeyBase   string
	VaultEncKey     string
	PGMetaCryptoKey string

	// Analytics (optional)
	LogflarePublicToken  string
	LogflarePrivateToken string

	// URLs (from config, not extracted)
	PublicURL string
	SiteURL   string
	APIURL    string
}

// Extractor orchestrates secret extraction from Supabase services
type Extractor struct {
	config *config.ExtractorConfig
	k8s    *k8s.Client
}

// NewExtractor creates a new Extractor instance
func NewExtractor(cfg *config.ExtractorConfig) (*Extractor, error) {
	k8sClient, err := k8s.NewClient(cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &Extractor{
		config: cfg,
		k8s:    k8sClient,
	}, nil
}

// Extract extracts all secrets from Supabase services
func (e *Extractor) Extract(ctx context.Context) (*ExtractedSecrets, error) {
	secrets := &ExtractedSecrets{
		PublicURL: e.config.PublicURL,
		SiteURL:   e.config.SiteURL,
	}

	// Set API URL (defaults to PublicURL if not specified)
	if e.config.PublicURL != "" {
		secrets.APIURL = e.config.PublicURL
	}

	// Wait for services to be ready
	log.Println("Waiting for Supabase services to be ready...")
	if err := e.waitForServices(ctx); err != nil {
		return nil, fmt.Errorf("services not ready: %w", err)
	}
	log.Println("Services are ready")

	// Extract from Postgres
	log.Println("Extracting secrets from Postgres...")
	pgSecrets, err := e.extractFromPostgres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract from postgres: %w", err)
	}
	secrets.PostgresPassword = pgSecrets.Password
	if pgSecrets.JWTSecret != "" {
		secrets.JWTSecret = pgSecrets.JWTSecret
	}
	log.Println("✓ Postgres secrets extracted")

	// Extract from Kong
	log.Println("Extracting secrets from Kong...")
	kongSecrets, err := e.extractFromKong(ctx)
	if err != nil {
		log.Printf("Warning: failed to extract from kong: %v", err)
	} else {
		if kongSecrets.AnonKey != "" {
			secrets.AnonKey = kongSecrets.AnonKey
		}
		if kongSecrets.ServiceRoleKey != "" {
			secrets.ServiceRoleKey = kongSecrets.ServiceRoleKey
		}
		if kongSecrets.DashboardUsername != "" {
			secrets.DashboardUsername = kongSecrets.DashboardUsername
		}
		if kongSecrets.DashboardPassword != "" {
			secrets.DashboardPassword = kongSecrets.DashboardPassword
		}
		log.Println("✓ Kong secrets extracted")
	}

	// Extract from Auth service
	log.Println("Extracting secrets from Auth service...")
	authSecrets, err := e.extractFromAuth(ctx)
	if err != nil {
		log.Printf("Warning: failed to extract from auth: %v", err)
	} else {
		if authSecrets.JWTSecret != "" && secrets.JWTSecret == "" {
			secrets.JWTSecret = authSecrets.JWTSecret
		}
		if authSecrets.SecretKeyBase != "" {
			secrets.SecretKeyBase = authSecrets.SecretKeyBase
		}
		if authSecrets.AnonKey != "" && secrets.AnonKey == "" {
			secrets.AnonKey = authSecrets.AnonKey
		}
		if authSecrets.ServiceKey != "" && secrets.ServiceRoleKey == "" {
			secrets.ServiceRoleKey = authSecrets.ServiceKey
		}
		log.Println("✓ Auth secrets extracted")
	}

	// Extract from Meta service (optional)
	log.Println("Extracting secrets from Meta service...")
	metaSecrets, err := e.extractFromMeta(ctx)
	if err != nil {
		log.Printf("Warning: failed to extract from meta: %v", err)
	} else {
		if cryptoKey, ok := metaSecrets["PG_META_CRYPTO_KEY"]; ok {
			secrets.PGMetaCryptoKey = cryptoKey
		}
		log.Println("✓ Meta secrets extracted")
	}

	// Validate all required secrets are present
	if err := e.validate(secrets); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	log.Println("✓ All secrets extracted and validated")
	return secrets, nil
}

// waitForServices waits for all required Supabase services to be ready
func (e *Extractor) waitForServices(ctx context.Context) error {
	requiredServices := []string{
		fmt.Sprintf("app=%s-db", e.config.ServicePrefix),
		fmt.Sprintf("app=%s-kong", e.config.ServicePrefix),
		fmt.Sprintf("app=%s-auth", e.config.ServicePrefix),
	}

	for _, selector := range requiredServices {
		log.Printf("Waiting for pods with selector: %s", selector)
		if err := e.k8s.WaitForPodsReady(ctx, selector, e.config.ReadyTimeout); err != nil {
			// Try alternative label format
			altSelector := strings.Replace(selector, "app=", "app.kubernetes.io/component=", 1)
			log.Printf("Trying alternative selector: %s", altSelector)
			if err := e.k8s.WaitForPodsReady(ctx, altSelector, e.config.ReadyTimeout); err != nil {
				return fmt.Errorf("service not ready (%s): %w", selector, err)
			}
		}
	}

	return nil
}

// PersistSecrets persists extracted secrets to a Kubernetes Secret
func (e *Extractor) PersistSecrets(ctx context.Context, secretName string, secrets *ExtractedSecrets) error {
	secretData := make(map[string]string)

	// Add all non-empty secrets
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
	if secrets.DashboardUsername != "" {
		secretData["DASHBOARD_USERNAME"] = secrets.DashboardUsername
	}
	if secrets.DashboardPassword != "" {
		secretData["DASHBOARD_PASSWORD"] = secrets.DashboardPassword
	}
	if secrets.SecretKeyBase != "" {
		secretData["SECRET_KEY_BASE"] = secrets.SecretKeyBase
	}
	if secrets.VaultEncKey != "" {
		secretData["VAULT_ENC_KEY"] = secrets.VaultEncKey
	}
	if secrets.PGMetaCryptoKey != "" {
		secretData["PG_META_CRYPTO_KEY"] = secrets.PGMetaCryptoKey
	}
	if secrets.LogflarePublicToken != "" {
		secretData["LOGFLARE_PUBLIC_ACCESS_TOKEN"] = secrets.LogflarePublicToken
	}
	if secrets.LogflarePrivateToken != "" {
		secretData["LOGFLARE_PRIVATE_ACCESS_TOKEN"] = secrets.LogflarePrivateToken
	}
	if secrets.PublicURL != "" {
		secretData["SUPABASE_PUBLIC_URL"] = secrets.PublicURL
		secretData["API_EXTERNAL_URL"] = secrets.PublicURL
	}
	if secrets.SiteURL != "" {
		secretData["SITE_URL"] = secrets.SiteURL
	}

	return e.k8s.PersistExtractedSecrets(ctx, secretName, secretData)
}

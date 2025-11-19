package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/config"
	"github.com/mjasion/k8s-supabase/secret-extractor/pkg/extractor"
)

func main() {
	// Parse command-line flags
	var (
		namespace      = flag.String("namespace", "default", "Kubernetes namespace")
		secretName     = flag.String("secret-name", "supabase-extracted-secrets", "Name of secret to create")
		servicePrefix  = flag.String("service-prefix", "supabase", "Prefix for service names")
		postgresHost   = flag.String("postgres-host", "postgres", "Postgres service host")
		postgresPort   = flag.Int("postgres-port", 5432, "Postgres port")
		postgresUser   = flag.String("postgres-user", "postgres", "Postgres user")
		publicURL      = flag.String("public-url", "", "Public URL (from config)")
		siteURL        = flag.String("site-url", "", "Site URL (from config)")
		retries        = flag.Int("retries", 10, "Number of retries")
		retryInterval  = flag.Duration("retry-interval", 10*time.Second, "Retry interval")
		readyTimeout   = flag.Duration("ready-timeout", 5*time.Minute, "Service ready timeout")
		kongConfigPath = flag.String("kong-config-path", "/home/kong/kong.yml", "Path to Kong config file")
	)

	flag.Parse()

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Supabase Secret Extractor")
	log.Printf("Namespace: %s", *namespace)
	log.Printf("Service Prefix: %s", *servicePrefix)
	log.Printf("Secret Name: %s", *secretName)

	// Create configuration
	cfg := &config.ExtractorConfig{
		Namespace:      *namespace,
		ServicePrefix:  *servicePrefix,
		SecretName:     *secretName,
		PostgresHost:   *postgresHost,
		PostgresPort:   *postgresPort,
		PostgresUser:   *postgresUser,
		Retries:        *retries,
		RetryInterval:  *retryInterval,
		ReadyTimeout:   *readyTimeout,
		PublicURL:      *publicURL,
		SiteURL:        *siteURL,
		KongConfigPath: *kongConfigPath,
	}

	// Create extractor
	e, err := extractor.NewExtractor(cfg)
	if err != nil {
		log.Fatalf("Failed to create extractor: %v", err)
	}

	ctx := context.Background()

	// Retry logic - wait for services to be ready and extract secrets
	var secrets *extractor.ExtractedSecrets
	var lastErr error

	log.Printf("Will attempt extraction up to %d times with %v interval", *retries, *retryInterval)

	for i := 0; i < *retries; i++ {
		log.Printf("Extraction attempt %d/%d", i+1, *retries)

		secrets, err = e.Extract(ctx)
		if err == nil {
			log.Println("✓ Secret extraction successful")
			break
		}

		lastErr = err
		log.Printf("Attempt %d/%d failed: %v", i+1, *retries, err)

		if i < *retries-1 {
			log.Printf("Retrying in %v...", *retryInterval)
			time.Sleep(*retryInterval)
		}
	}

	if lastErr != nil {
		log.Fatalf("Failed to extract secrets after %d attempts: %v", *retries, lastErr)
	}

	// Persist secrets to Kubernetes
	log.Printf("Persisting secrets to Kubernetes Secret: %s", *secretName)
	if err := e.PersistSecrets(ctx, *secretName, secrets); err != nil {
		log.Fatalf("Failed to persist secrets: %v", err)
	}

	log.Printf("✓ Successfully persisted secrets to '%s'", *secretName)

	// Print summary
	printSecretsSummary(secrets)

	// Print usage instructions
	printUsageInstructions(*namespace, *secretName)

	log.Println("Secret extraction completed successfully")
	os.Exit(0)
}

func printSecretsSummary(secrets *extractor.ExtractedSecrets) {
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("                  SUPABASE SECRETS SUMMARY")
	fmt.Println("================================================================")
	fmt.Println()

	if secrets.DashboardUsername != "" && secrets.DashboardPassword != "" {
		fmt.Println("Dashboard Credentials:")
		fmt.Printf("  Username: %s\n", secrets.DashboardUsername)
		fmt.Printf("  Password: %s\n", secrets.DashboardPassword)
		fmt.Println()
	}

	if secrets.AnonKey != "" {
		fmt.Println("API Keys (for client applications):")
		if len(secrets.AnonKey) > 50 {
			fmt.Printf("  Anon Key: %s...%s\n", secrets.AnonKey[:30], secrets.AnonKey[len(secrets.AnonKey)-20:])
		} else {
			fmt.Printf("  Anon Key: %s\n", secrets.AnonKey)
		}
		if len(secrets.ServiceRoleKey) > 50 {
			fmt.Printf("  Service Role Key: %s...%s\n", secrets.ServiceRoleKey[:30], secrets.ServiceRoleKey[len(secrets.ServiceRoleKey)-20:])
		} else {
			fmt.Printf("  Service Role Key: %s\n", secrets.ServiceRoleKey)
		}
		fmt.Println()
	}

	if secrets.PublicURL != "" {
		fmt.Println("URLs:")
		fmt.Printf("  Public URL: %s\n", secrets.PublicURL)
		fmt.Printf("  Site URL: %s\n", secrets.SiteURL)
		fmt.Println()
	}

	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("IMPORTANT:")
	fmt.Println("  - Save these credentials securely")
	fmt.Println("  - Update default passwords before production use")
	fmt.Println("  - Keep the JWT secret and service role key confidential")
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println()
}

func printUsageInstructions(namespace, secretName string) {
	fmt.Println("Usage Instructions:")
	fmt.Println()
	fmt.Println("To view the secret in Kubernetes:")
	fmt.Printf("  kubectl get secret %s -n %s -o yaml\n", secretName, namespace)
	fmt.Println()
	fmt.Println("To use in client applications:")
	fmt.Println("  Add the following to your deployment:")
	fmt.Println()
	fmt.Println("  env:")
	fmt.Println("    - name: SUPABASE_URL")
	fmt.Println("      valueFrom:")
	fmt.Println("        secretKeyRef:")
	fmt.Printf("          name: %s\n", secretName)
	fmt.Println("          key: SUPABASE_PUBLIC_URL")
	fmt.Println("    - name: SUPABASE_ANON_KEY")
	fmt.Println("      valueFrom:")
	fmt.Println("        secretKeyRef:")
	fmt.Printf("          name: %s\n", secretName)
	fmt.Println("          key: ANON_KEY")
	fmt.Println()
	fmt.Println("To decode a secret value:")
	fmt.Printf("  kubectl get secret %s -n %s -o jsonpath='{.data.ANON_KEY}' | base64 -d\n", secretName, namespace)
	fmt.Println()
}

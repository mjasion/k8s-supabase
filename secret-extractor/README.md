# Supabase Secret Extractor

A Kubernetes-native service that extracts secrets from running Supabase containers and persists them to Kubernetes Secret resources for easy client access and redeployment.

## Overview

The Supabase Secret Extractor solves the problem of secret management in Kubernetes deployments by:

1. **Extracting secrets** from running Supabase containers (Postgres, Kong, Auth, etc.)
2. **Validating** that all required secrets are present and correctly formatted
3. **Persisting** secrets to a Kubernetes Secret resource
4. **Enabling** easy client access and secret reuse across deployments

## Architecture

```
┌────────────────────────────────────────────────────────┐
│ Supabase Services Running                              │
│ - Postgres (generates passwords)                       │
│ - Kong (stores API keys)                               │
│ - Auth (manages JWT secrets)                           │
└───────────────────┬────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────────┐
│ Secret Extractor Job                                   │
│ - Waits for services to be ready                       │
│ - Execs into pods to read environment variables        │
│ - Queries Postgres database                            │
│ - Reads Kong configuration files                       │
│ - Validates all extracted secrets                      │
└───────────────────┬────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────────┐
│ Kubernetes Secret: supabase-extracted-secrets          │
│ - POSTGRES_PASSWORD                                    │
│ - JWT_SECRET                                           │
│ - ANON_KEY                                             │
│ - SERVICE_ROLE_KEY                                     │
│ - DASHBOARD_USERNAME / PASSWORD                        │
│ - ... and more                                         │
└────────────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────────┐
│ Client Applications                                    │
│ - Read secrets from Kubernetes Secret                  │
│ - Use ANON_KEY for client-side operations             │
│ - Use SERVICE_ROLE_KEY for server-side operations     │
└────────────────────────────────────────────────────────┘
```

## Features

- ✅ **Non-invasive**: Reads from running containers without modifying them
- ✅ **Automatic**: Runs as a Kubernetes Job after Supabase deployment
- ✅ **Resilient**: Retries and waits for services to be ready
- ✅ **Validated**: Ensures all required secrets are present and correctly formatted
- ✅ **Kubernetes-native**: Stores secrets in Kubernetes Secret resources
- ✅ **Client-friendly**: Easy access for client applications

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/mjasion/k8s-supabase.git
cd k8s-supabase/secret-extractor

# Install dependencies
go mod download

# Build the binary
make build

# Or build Docker image
make docker-build
```

### Use Pre-built Docker Image

```bash
docker pull ghcr.io/mjasion/supabase-secret-extractor:latest
```

## Usage

### As a Kubernetes Job (Recommended)

The extractor is designed to run as a post-install Kubernetes Job:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: supabase-secret-extractor
spec:
  template:
    spec:
      serviceAccountName: supabase-extractor
      restartPolicy: OnFailure
      containers:
      - name: extractor
        image: ghcr.io/mjasion/supabase-secret-extractor:latest
        args:
          - --namespace=supabase
          - --service-prefix=supabase
          - --secret-name=supabase-extracted-secrets
          - --postgres-host=supabase-db
          - --public-url=https://api.yourdomain.com
          - --site-url=https://app.yourdomain.com
          - --retries=10
          - --retry-interval=10s
```

### Standalone Usage

```bash
./secret-extractor \
  --namespace=supabase \
  --service-prefix=supabase \
  --secret-name=supabase-extracted-secrets \
  --postgres-host=supabase-db \
  --postgres-port=5432 \
  --public-url=https://api.yourdomain.com \
  --site-url=https://app.yourdomain.com
```

### Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `--namespace` | `default` | Kubernetes namespace where Supabase is deployed |
| `--service-prefix` | `supabase` | Prefix used for Supabase service names |
| `--secret-name` | `supabase-extracted-secrets` | Name of the Kubernetes Secret to create |
| `--postgres-host` | `postgres` | Postgres service hostname |
| `--postgres-port` | `5432` | Postgres port |
| `--postgres-user` | `postgres` | Postgres username |
| `--public-url` | *(required)* | Public-facing URL for the Supabase API |
| `--site-url` | *(required)* | Site URL for email templates and redirects |
| `--retries` | `10` | Number of extraction attempts |
| `--retry-interval` | `10s` | Time to wait between retries |
| `--ready-timeout` | `5m` | Timeout for waiting for services to be ready |
| `--kong-config-path` | `/home/kong/kong.yml` | Path to Kong config file in the pod |

## Extracted Secrets

The extractor retrieves the following secrets:

### Core Secrets
- **POSTGRES_PASSWORD**: PostgreSQL database password
- **JWT_SECRET**: Secret used to sign JWT tokens
- **ANON_KEY**: Anonymous/public JWT token for client use
- **SERVICE_ROLE_KEY**: Service role JWT token for server-side operations

### Dashboard
- **DASHBOARD_USERNAME**: Studio dashboard username
- **DASHBOARD_PASSWORD**: Studio dashboard password

### Encryption Keys
- **SECRET_KEY_BASE**: GoTrue secret key base
- **PG_META_CRYPTO_KEY**: Database metadata encryption key
- **VAULT_ENC_KEY**: Vault encryption key (if available)

### URLs
- **SUPABASE_PUBLIC_URL**: Public API URL
- **API_EXTERNAL_URL**: External API URL (same as public URL)
- **SITE_URL**: Site URL for redirects

## RBAC Requirements

The extractor needs specific Kubernetes permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: supabase-extractor
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "get", "update", "patch"]
- apiGroups: [""]
  resources: ["pods", "pods/exec"]
  verbs: ["get", "list", "create"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
```

## Using Extracted Secrets in Client Apps

### Environment Variables

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: my-app:latest
        env:
          - name: SUPABASE_URL
            valueFrom:
              secretKeyRef:
                name: supabase-extracted-secrets
                key: SUPABASE_PUBLIC_URL
          - name: SUPABASE_ANON_KEY
            valueFrom:
              secretKeyRef:
                name: supabase-extracted-secrets
                key: ANON_KEY
          - name: SUPABASE_SERVICE_ROLE_KEY
            valueFrom:
              secretKeyRef:
                name: supabase-extracted-secrets
                key: SERVICE_ROLE_KEY
```

### ConfigMap Reference

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.json: |
    {
      "supabaseUrl": "$(kubectl get secret supabase-extracted-secrets -o jsonpath='{.data.SUPABASE_PUBLIC_URL}' | base64 -d)",
      "supabaseAnonKey": "$(kubectl get secret supabase-extracted-secrets -o jsonpath='{.data.ANON_KEY}' | base64 -d)"
    }
```

## Extraction Process

### 1. Wait for Services
The extractor waits for all required Supabase services to be ready:
- Postgres (`db`)
- Kong (`kong`)
- Auth (`auth`)

### 2. Extract from Postgres
- Connects to Postgres using environment password
- Queries for JWT secret from database settings
- Validates connection

### 3. Extract from Kong
- Reads Kong declarative config (`/home/kong/kong.yml`)
- Parses consumer credentials for API keys
- Falls back to environment variables if config not available

### 4. Extract from Auth
- Reads environment variables from Auth (GoTrue) pod
- Extracts JWT secret, secret key base
- Retrieves API keys if present

### 5. Extract from Other Services
- Meta: crypto keys
- Storage: additional credentials
- Analytics: access tokens (optional)

### 6. Validate
- Checks all required secrets are present
- Validates JWT format for API keys
- Ensures minimum secret lengths
- Validates JWT secret is at least 32 characters

### 7. Persist
- Creates or updates Kubernetes Secret
- Adds metadata labels and annotations
- Outputs summary to logs

## Troubleshooting

### Extraction Fails - Services Not Ready

```bash
# Check if all pods are running
kubectl get pods -n supabase

# Check extractor logs
kubectl logs job/supabase-secret-extractor -n supabase

# Solution: Increase retries or retry interval
--retries=20 --retry-interval=15s
```

### Secrets Not Found in Pods

```bash
# Verify secrets exist in target pod
kubectl exec -n supabase supabase-db-0 -- env | grep POSTGRES_PASSWORD

# Check RBAC permissions
kubectl describe role supabase-extractor -n supabase
kubectl describe rolebinding supabase-extractor -n supabase
```

### Connection to Postgres Fails

```bash
# Test postgres connection manually
kubectl exec -n supabase supabase-db-0 -- psql -U postgres -c "SELECT 1"

# Check postgres service
kubectl get svc -n supabase | grep postgres

# Verify network policies aren't blocking
kubectl get networkpolicies -n supabase
```

### Incomplete Secret Extraction

```bash
# View extracted secret
kubectl get secret supabase-extracted-secrets -n supabase -o yaml

# Check which keys are missing
kubectl get secret supabase-extracted-secrets -n supabase -o jsonpath='{.data}' | jq 'keys'

# Re-run extractor with verbose logging
kubectl logs job/supabase-secret-extractor -n supabase -f
```

## Development

### Running Tests

```bash
make test
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Build and push
make docker-push
```

### Local Development

```bash
# Run against local Kubernetes cluster
make run

# Or with custom settings
./secret-extractor \
  --namespace=supabase-dev \
  --service-prefix=supabase \
  --public-url=http://localhost:8000 \
  --site-url=http://localhost:3000
```

## Security Considerations

### Principle of Least Privilege
- Extractor only has permissions to read specific pods and create/update one secret
- Runs as non-root user in container
- Uses service account with minimal RBAC

### Secret Protection
- Secrets are never logged in full (truncated for display)
- Kubernetes Secret is encrypted at rest (by Kubernetes)
- Recommend using encryption providers for additional security

### Audit Trail
- Job logs contain extraction timestamps
- Secret metadata includes extraction time annotation
- All Kubernetes API calls are logged by apiserver

## Integration with Helm

The extractor integrates seamlessly with the Supabase Helm chart:

```yaml
# values.yaml
secretExtractor:
  enabled: true
  image:
    repository: ghcr.io/mjasion/supabase-secret-extractor
    tag: latest
  secretName: supabase-extracted-secrets
  retries: 10
  retryInterval: "10s"
```

## License

This project is part of the k8s-supabase repository.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

- Issues: https://github.com/mjasion/k8s-supabase/issues
- Documentation: https://github.com/mjasion/k8s-supabase/tree/main/docs

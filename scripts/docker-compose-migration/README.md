# Supabase Docker Compose to Kubernetes Migration Tool

This tool automatically converts Supabase's official docker-compose configuration to Kubernetes manifests in both Kustomize and Helm chart formats.

## Features

- ✅ Fetches the latest Supabase docker-compose configuration from GitHub
- ✅ Converts all services to Kubernetes manifests (Deployments, StatefulSets, Services)
- ✅ Generates Kustomize base and overlays (development, staging, production)
- ✅ Generates complete Helm chart with customizable values
- ✅ Handles secrets, ConfigMaps, and environment variables
- ✅ Converts health checks to Kubernetes probes
- ✅ Supports persistent volumes for stateful services
- ✅ Includes post-install job to extract and display secrets

## Prerequisites

- Go 1.21 or higher
- Internet connection (to fetch from GitHub)

## Installation

```bash
cd scripts/docker-compose-migration
go mod download
go build -o supabase-converter .
```

## Usage

### Basic Usage (Fetch from GitHub)

```bash
./supabase-converter
```

This will:
1. Fetch the latest docker-compose.yml from https://github.com/supabase/supabase
2. Generate Kustomize manifests in `../../kustomize/`
3. Generate Helm chart in `../../charts/supabase/`

### Advanced Options

```bash
# Use a specific branch
./supabase-converter -branch develop

# Use a different repository
./supabase-converter -repo https://github.com/your-fork/supabase

# Use local docker-compose files
./supabase-converter -skip-fetch -local /path/to/docker/directory

# Specify custom output directory
./supabase-converter -output /path/to/output
```

### Available Flags

- `-repo`: Supabase repository URL (default: https://github.com/supabase/supabase)
- `-branch`: Git branch to fetch from (default: master)
- `-docker-path`: Path to docker directory in repo (default: docker)
- `-output`: Output directory for generated files (default: ../../)
- `-skip-fetch`: Skip fetching from GitHub
- `-local`: Path to local docker-compose directory (requires -skip-fetch)

## Generated Structure

After running the converter, you'll have:

```
k8s-supabase/
├── kustomize/
│   ├── base/
│   │   ├── kustomization.yaml
│   │   ├── namespace.yaml
│   │   ├── configmap-*.yaml
│   │   ├── secret-*.yaml
│   │   ├── deployment-*.yaml
│   │   ├── statefulset-*.yaml
│   │   ├── service-*.yaml
│   │   └── secret-extractor-*.yaml
│   └── overlays/
│       ├── development/
│       ├── staging/
│       └── production/
│
└── charts/
    └── supabase/
        ├── Chart.yaml
        ├── values.yaml
        ├── templates/
        │   ├── _helpers.tpl
        │   ├── NOTES.txt
        │   ├── namespace.yaml
        │   ├── configmap-*.yaml
        │   ├── secret-*.yaml
        │   ├── serviceaccount.yaml
        │   ├── secret-extractor-job.yaml
        │   ├── secret-extractor-rbac.yaml
        │   └── [service-directories]/
        │       ├── deployment.yaml
        │       ├── statefulset.yaml
        │       └── service.yaml
        └── ...
```

## Deployment

### Using Kustomize

```bash
# Deploy base configuration
kubectl apply -k kustomize/base

# Deploy with overlay
kubectl apply -k kustomize/overlays/production
```

### Using Helm

```bash
# Install with default values
helm install supabase charts/supabase -n supabase --create-namespace

# Install with custom values
helm install supabase charts/supabase -n supabase --create-namespace -f custom-values.yaml

# Upgrade
helm upgrade supabase charts/supabase -n supabase
```

## Post-Deployment

After deployment, a post-install job will automatically run to extract and display your secrets. View the job logs:

```bash
kubectl logs -n supabase job/supabase-secret-extractor
```

You'll see:
- Dashboard credentials (username/password)
- JWT tokens (anon key, service role key)
- Database password

## Configuration

### Updating Secrets

**IMPORTANT**: Before deploying to production, update all secrets in:

**Kustomize**: Edit `kustomize/base/secret-*.yaml` files

**Helm**: Edit `charts/supabase/values.yaml`:

```yaml
secrets:
  jwt:
    secret: "your-super-secret-jwt-token-with-at-least-32-characters-long"
    anonKey: "your-anon-jwt-token"
    serviceRoleKey: "your-service-role-jwt-token"
  database:
    password: "your-super-secret-and-long-postgres-password"
  dashboard:
    username: "admin"
    password: "secure-password"
```

### Scaling Services

**Kustomize**: Add replicas patch in overlays

**Helm**: Update `values.yaml`:

```yaml
kong:
  replicas: 2
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
```

## Services Converted

The tool converts all Supabase services:

- 🗄️ **PostgreSQL** (db) - Database
- 🌐 **Kong** - API Gateway
- 🔐 **GoTrue** (auth) - Authentication
- 📡 **PostgREST** (rest) - REST API
- ⚡ **Realtime** - Real-time subscriptions
- 📦 **Storage** - File storage
- 🖼️ **Imgproxy** - Image processing
- 🔧 **Meta** - Database metadata
- 🎨 **Studio** - Dashboard UI
- ⚙️ **Edge Functions** (functions) - Serverless functions
- 📊 **Analytics** - Logging and analytics
- 📈 **Vector** - Log aggregation
- 🔌 **Supavisor** - Connection pooler

## Troubleshooting

### Connection Issues

```bash
# Check pod status
kubectl get pods -n supabase

# View logs
kubectl logs -n supabase <pod-name>

# Describe resources
kubectl describe pod -n supabase <pod-name>
```

### Secret Access

```bash
# View secrets
kubectl get secrets -n supabase

# Decode a specific secret
kubectl get secret supabase-jwt -n supabase -o jsonpath='{.data.ANON_KEY}' | base64 -d
```

### Port Forwarding

```bash
# Studio (Dashboard)
kubectl port-forward -n supabase svc/supabase-studio 3000:3000

# Kong (API Gateway)
kubectl port-forward -n supabase svc/supabase-kong 8000:8000

# PostgreSQL
kubectl port-forward -n supabase svc/supabase-db 5432:5432
```

## Development

### Project Structure

```
docker-compose-migration/
├── main.go                 # Entry point
├── pkg/
│   ├── fetcher/           # GitHub file fetcher
│   ├── parser/            # Docker Compose parser
│   ├── converter/         # Kubernetes converter
│   ├── kustomize/         # Kustomize generator
│   └── helm/              # Helm chart generator
├── go.mod
├── go.sum
└── README.md
```

### Adding New Services

The tool automatically detects and converts all services from the docker-compose file. No code changes needed for new Supabase services.

### Customizing Templates

**Kustomize**: Modify generators in `pkg/kustomize/generator.go`

**Helm**: Modify templates in `pkg/helm/generator.go`

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This tool is provided as-is for converting Supabase configurations to Kubernetes.

## Resources

- [Supabase Documentation](https://supabase.com/docs)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [Helm Documentation](https://helm.sh/docs/)

## Support

For issues related to:
- **This tool**: Open an issue in this repository
- **Supabase**: Visit [Supabase GitHub](https://github.com/supabase/supabase)
- **Kubernetes**: Visit [Kubernetes Community](https://kubernetes.io/community/)

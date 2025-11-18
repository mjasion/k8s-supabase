# Supabase Kubernetes Deployment

This repository contains Kubernetes manifests and Helm charts for deploying Supabase on Kubernetes.

## 🚀 Quick Start

### Using the Migration Tool

Convert Supabase's docker-compose configuration to Kubernetes manifests:

```bash
cd scripts/docker-compose-migration
make build
./supabase-converter
```

This will generate:
- **Kustomize** manifests in `kustomize/`
- **Helm chart** in `charts/supabase/`

### Deploy with Helm

```bash
helm install supabase charts/supabase -n supabase --create-namespace
```

### Deploy with Kustomize

```bash
kubectl apply -k kustomize/base
```

## 📁 Repository Structure

```
k8s-supabase/
├── README.md
├── .gitignore
│
├── charts/
│   └── supabase/                   # Helm chart for Supabase
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── templates/
│       │   ├── _helpers.tpl
│       │   ├── NOTES.txt
│       │   ├── configmap.yaml
│       │   ├── secret-extractor-job.yaml
│       │   ├── secret-extractor-rbac.yaml
│       │   ├── postgres/
│       │   ├── kong/
│       │   ├── auth/
│       │   ├── rest/
│       │   ├── realtime/
│       │   ├── storage/
│       │   ├── meta/
│       │   ├── studio/
│       │   ├── functions/
│       │   ├── analytics/
│       │   ├── imgproxy/
│       │   └── vector/
│       └── crds/
│
├── kustomize/
│   ├── base/                       # Base Kubernetes manifests
│   │   ├── kustomization.yaml
│   │   ├── namespace.yaml
│   │   ├── secret-extractor-job.yaml
│   │   ├── postgres/
│   │   ├── kong/
│   │   └── ... (all services)
│   │
│   └── overlays/                   # Environment-specific overlays
│       ├── development/
│       ├── staging/
│       └── production/
│
├── scripts/
│   └── docker-compose-migration/   # Migration tool
│       ├── main.go
│       ├── pkg/
│       ├── Makefile
│       └── README.md
│
└── docs/
    ├── architecture.md
    ├── docker-compose-migration.md
    ├── secret-extraction.md
    ├── helm-installation.md
    ├── kustomize-installation.md
    └── troubleshooting.md
```

## 🔧 Features

- ✅ **Automated Conversion**: Go-based tool to convert docker-compose to Kubernetes
- ✅ **Helm Chart**: Production-ready Helm chart with customizable values
- ✅ **Kustomize**: Declarative configuration with overlays for different environments
- ✅ **Secret Extraction**: Advanced Go service that extracts secrets from running Supabase containers
- ✅ **Secret Management**: Secure handling of JWT tokens, passwords, and API keys
- ✅ **Client Integration**: Easy secret access for client applications via Kubernetes Secrets
- ✅ **StatefulSets**: Proper handling of stateful services (PostgreSQL, Storage)
- ✅ **Health Checks**: Kubernetes liveness and readiness probes
- ✅ **Resource Management**: Configurable CPU and memory limits
- ✅ **Multi-Environment**: Support for dev, staging, and production

## 🔑 Secret Extraction

This repository includes a sophisticated **Secret Extractor** service that automatically extracts secrets from running Supabase containers and persists them to Kubernetes Secrets.

### How It Works

1. **Waits for Services**: Ensures all Supabase services are ready
2. **Extracts Secrets**: Reads from Postgres, Kong, Auth, and other services
3. **Validates**: Ensures all required secrets are present and correctly formatted
4. **Persists**: Stores secrets in a Kubernetes Secret resource
5. **Enables Access**: Client apps can easily reference the Secret

### Extracted Secrets

- `POSTGRES_PASSWORD` - Database password
- `JWT_SECRET` - JWT signing secret
- `ANON_KEY` - Anonymous/public API key
- `SERVICE_ROLE_KEY` - Service role API key
- `DASHBOARD_USERNAME` / `DASHBOARD_PASSWORD` - Studio credentials
- URLs and configuration

See [secret-extractor/README.md](secret-extractor/README.md) for detailed documentation.

## 📦 Components

Supabase consists of the following services:

| Service | Description | Type |
|---------|-------------|------|
| **PostgreSQL** (db) | Main database | StatefulSet |
| **Kong** | API Gateway | Deployment |
| **GoTrue** (auth) | Authentication service | Deployment |
| **PostgREST** (rest) | Auto-generated REST API | Deployment |
| **Realtime** | WebSocket connections | Deployment |
| **Storage** | File storage API | StatefulSet |
| **Imgproxy** | Image transformation | Deployment |
| **Meta** | Database metadata API | Deployment |
| **Studio** | Web dashboard | Deployment |
| **Edge Functions** (functions) | Serverless functions | Deployment |
| **Analytics** | Logging and analytics | Deployment |
| **Vector** | Log aggregation | Deployment |
| **Supavisor** | Connection pooler | Deployment |

## 🔐 Security

**⚠️ IMPORTANT**: Before deploying to production:

1. **Update all secrets** in `charts/supabase/values.yaml` or `kustomize/base/secret-*.yaml`
2. **Generate new JWT tokens** using the Supabase CLI or online tools
3. **Change default passwords** for PostgreSQL and dashboard
4. **Enable TLS/SSL** for external access
5. **Configure ingress** with proper authentication
6. **Set up backup** procedures for your database

See [docs/secret-extraction.md](docs/secret-extraction.md) for detailed guidance.

## 📖 Documentation

- [Migration Guide](docs/docker-compose-migration.md) - Convert docker-compose to Kubernetes
- [Helm Installation](docs/helm-installation.md) - Deploy using Helm
- [Kustomize Installation](docs/kustomize-installation.md) - Deploy using Kustomize
- [Architecture Overview](docs/architecture.md) - System architecture and design
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

## 🛠️ Development

### Prerequisites

- Go 1.21+ (for migration tool)
- kubectl
- Helm 3.x (for Helm deployment)
- Kustomize (for Kustomize deployment)

### Building the Migration Tool

```bash
cd scripts/docker-compose-migration
make build
```

### Running Tests

```bash
cd scripts/docker-compose-migration
make test
```

### Regenerating Manifests

```bash
cd scripts/docker-compose-migration
make clean generate
```

## 🚀 Deployment Options

### Option 1: Helm (Recommended)

```bash
# Install with default values
helm install supabase charts/supabase -n supabase --create-namespace

# Install with custom values
helm install supabase charts/supabase -n supabase --create-namespace -f my-values.yaml

# Upgrade
helm upgrade supabase charts/supabase -n supabase
```

### Option 2: Kustomize

```bash
# Development
kubectl apply -k kustomize/overlays/development

# Production
kubectl apply -k kustomize/overlays/production
```

## 🔍 Accessing Services

### Port Forwarding (Development)

```bash
# Studio Dashboard
kubectl port-forward -n supabase svc/supabase-studio 3000:3000
# Visit: http://localhost:3000

# API Gateway
kubectl port-forward -n supabase svc/supabase-kong 8000:8000
# API: http://localhost:8000

# PostgreSQL
kubectl port-forward -n supabase svc/supabase-db 5432:5432
```

### Production Access

Set up an Ingress controller and configure routes for:
- Studio: `studio.your-domain.com`
- API: `api.your-domain.com`

## 📊 Monitoring

After deployment, check the status:

```bash
# Check pods
kubectl get pods -n supabase

# Check services
kubectl get svc -n supabase

# View logs
kubectl logs -n supabase -l app=supabase-kong

# View secrets (post-install job)
kubectl logs -n supabase job/supabase-secret-extractor
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is provided as-is for deploying Supabase on Kubernetes.

## 🔗 Resources

- [Supabase](https://supabase.com/) - Official website
- [Supabase Documentation](https://supabase.com/docs) - Official docs
- [Supabase GitHub](https://github.com/supabase/supabase) - Source code
- [Kubernetes](https://kubernetes.io/) - Container orchestration
- [Helm](https://helm.sh/) - Package manager for Kubernetes
- [Kustomize](https://kustomize.io/) - Kubernetes configuration management

## 💬 Support

- For issues with this repository: Open an issue here
- For Supabase questions: Visit [Supabase Community](https://github.com/supabase/supabase/discussions)
- For Kubernetes help: Visit [Kubernetes Community](https://kubernetes.io/community/)

---

**Made with ❤️ for the Supabase and Kubernetes communities**

# Docker Compose to Kubernetes Migration Guide

This guide explains how to use the migration tool to convert Supabase's docker-compose configuration to Kubernetes manifests.

## Overview

The migration tool automates the conversion of Supabase's official docker-compose setup into production-ready Kubernetes manifests. It generates both:

1. **Kustomize** manifests for declarative configuration management
2. **Helm charts** for templated, reusable deployments

## Migration Process

### Step 1: Run the Converter

Navigate to the migration script directory:

```bash
cd scripts/docker-compose-migration
```

Build and run the converter:

```bash
make build
./supabase-converter
```

The tool will:
1. Fetch `docker-compose.yml` from the Supabase GitHub repository
2. Parse all service definitions
3. Convert services to Kubernetes resources
4. Generate Kustomize manifests
5. Generate Helm chart templates

### Step 2: Review Generated Files

#### Kustomize Structure

```
kustomize/
├── base/
│   ├── kustomization.yaml           # Base kustomization
│   ├── namespace.yaml               # Namespace definition
│   ├── configmap-supabase-config.yaml
│   ├── secret-supabase-secrets.yaml
│   ├── secret-supabase-jwt.yaml
│   ├── secret-supabase-db.yaml
│   ├── deployment-*.yaml            # Service deployments
│   ├── statefulset-*.yaml          # Stateful services
│   ├── service-*.yaml              # Service definitions
│   └── secret-extractor-*.yaml     # Post-deployment job
└── overlays/
    ├── development/
    │   └── kustomization.yaml
    ├── staging/
    │   └── kustomization.yaml
    └── production/
        └── kustomization.yaml
```

#### Helm Chart Structure

```
charts/supabase/
├── Chart.yaml                      # Chart metadata
├── values.yaml                     # Default configuration
├── templates/
│   ├── _helpers.tpl               # Template helpers
│   ├── NOTES.txt                  # Post-install notes
│   ├── namespace.yaml
│   ├── configmap-*.yaml
│   ├── secret-*.yaml
│   ├── serviceaccount.yaml
│   ├── secret-extractor-job.yaml
│   ├── secret-extractor-rbac.yaml
│   └── [services]/                # Per-service templates
│       ├── deployment.yaml
│       ├── statefulset.yaml
│       └── service.yaml
```

### Step 3: Update Secrets

**⚠️ CRITICAL**: The default secrets are NOT secure for production use.

#### For Kustomize

Edit the secret files in `kustomize/base/`:

```bash
# Edit JWT secrets
vim kustomize/base/secret-supabase-jwt.yaml

# Edit database secrets
vim kustomize/base/secret-supabase-db.yaml

# Edit dashboard secrets
vim kustomize/base/secret-supabase-secrets.yaml
```

#### For Helm

Edit `charts/supabase/values.yaml`:

```yaml
secrets:
  jwt:
    secret: "CHANGE-ME-32-characters-minimum"
    anonKey: "GENERATE-NEW-JWT-TOKEN"
    serviceRoleKey: "GENERATE-NEW-JWT-TOKEN"
  database:
    password: "CHANGE-ME-strong-password"
  dashboard:
    username: "admin"
    password: "CHANGE-ME-strong-password"
```

#### Generating JWT Tokens

You can generate JWT tokens using the Supabase CLI or online tools:

```bash
# Install Supabase CLI
npm install -g supabase

# Generate JWT tokens
supabase gen keys jwt --exp-time 315360000
```

Or use: https://supabase.com/docs/guides/self-hosting#api-keys

### Step 4: Configure Resources

#### Kustomize Resource Customization

Create patches in overlays. Example for production:

```yaml
# kustomize/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

replicas:
  - name: supabase-kong
    count: 3
  - name: supabase-auth
    count: 2

patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: supabase-kong
      spec:
        template:
          spec:
            containers:
            - name: kong
              resources:
                requests:
                  memory: "512Mi"
                  cpu: "250m"
                limits:
                  memory: "1Gi"
                  cpu: "500m"
    target:
      kind: Deployment
      name: supabase-kong
```

#### Helm Resource Customization

Edit `values.yaml` or create custom values files:

```yaml
# custom-production-values.yaml
db:
  enabled: true
  replicas: 1
  resources:
    requests:
      memory: "2Gi"
      cpu: "1000m"
    limits:
      memory: "4Gi"
      cpu: "2000m"
  persistence:
    enabled: true
    storageClass: "fast-ssd"
    size: "50Gi"

kong:
  enabled: true
  replicas: 3
  service:
    type: LoadBalancer
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"

auth:
  enabled: true
  replicas: 2
```

### Step 5: Deploy

#### Option A: Deploy with Kustomize

```bash
# Development
kubectl apply -k kustomize/overlays/development

# Production
kubectl apply -k kustomize/overlays/production
```

#### Option B: Deploy with Helm

```bash
# Development
helm install supabase charts/supabase \
  -n supabase-dev \
  --create-namespace \
  -f charts/supabase/values.yaml

# Production
helm install supabase charts/supabase \
  -n supabase \
  --create-namespace \
  -f custom-production-values.yaml
```

### Step 6: Verify Deployment

Check pod status:

```bash
kubectl get pods -n supabase
```

All pods should be in `Running` state:

```
NAME                               READY   STATUS    RESTARTS   AGE
supabase-kong-xxx                  1/1     Running   0          2m
supabase-auth-xxx                  1/1     Running   0          2m
supabase-rest-xxx                  1/1     Running   0          2m
supabase-realtime-xxx              1/1     Running   0          2m
supabase-storage-xxx               1/1     Running   0          2m
supabase-db-0                      1/1     Running   0          2m
...
```

### Step 7: Access Secrets

After deployment, view the post-install job logs to see your credentials:

```bash
kubectl logs -n supabase job/supabase-secret-extractor
```

Output example:

```
================================================================
Supabase Secrets
================================================================

Dashboard Credentials:
  Username: admin
  Password: your-password

JWT Tokens:
  Anon Key: eyJhbGci...
  Service Role Key: eyJhbGci...

Database:
  Password: your-db-password

================================================================
```

### Step 8: Access Services

#### Port Forwarding (Development)

```bash
# Studio Dashboard
kubectl port-forward -n supabase svc/supabase-studio 3000:3000
# Access: http://localhost:3000

# API Gateway
kubectl port-forward -n supabase svc/supabase-kong 8000:8000
# Access: http://localhost:8000

# Database
kubectl port-forward -n supabase svc/supabase-db 5432:5432
# Access: postgresql://postgres:password@localhost:5432/postgres
```

#### Ingress (Production)

Create an Ingress resource:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: supabase-ingress
  namespace: supabase
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
    - hosts:
        - api.your-domain.com
        - studio.your-domain.com
      secretName: supabase-tls
  rules:
    - host: api.your-domain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: supabase-kong
                port:
                  number: 8000
    - host: studio.your-domain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: supabase-studio
                port:
                  number: 3000
```

## Migration Checklist

- [ ] Run the converter tool
- [ ] Review generated manifests
- [ ] Update all secrets (JWT, database, dashboard)
- [ ] Configure resource limits
- [ ] Set up persistent storage for database
- [ ] Configure ingress for external access
- [ ] Enable TLS/SSL certificates
- [ ] Set up monitoring and logging
- [ ] Configure backups for PostgreSQL
- [ ] Test all services
- [ ] Document custom configurations

## Service Mapping

| Docker Compose Service | Kubernetes Resource | Type |
|------------------------|---------------------|------|
| db | supabase-db | StatefulSet |
| kong | supabase-kong | Deployment |
| auth | supabase-auth | Deployment |
| rest | supabase-rest | Deployment |
| realtime | supabase-realtime | Deployment |
| storage | supabase-storage | StatefulSet |
| meta | supabase-meta | Deployment |
| studio | supabase-studio | Deployment |
| functions | supabase-functions | Deployment |
| analytics | supabase-analytics | Deployment |
| imgproxy | supabase-imgproxy | Deployment |
| vector | supabase-vector | Deployment |

## Differences from Docker Compose

### Networking

- Docker Compose: Services use container names for DNS
- Kubernetes: Services use Kubernetes Service DNS (e.g., `supabase-db.supabase.svc.cluster.local`)

### Storage

- Docker Compose: Named volumes
- Kubernetes: PersistentVolumeClaims with storage classes

### Environment Variables

- Docker Compose: `.env` file
- Kubernetes: ConfigMaps and Secrets

### Health Checks

- Docker Compose: Basic health checks
- Kubernetes: Liveness and readiness probes

### Scaling

- Docker Compose: Manual or Docker Swarm
- Kubernetes: Horizontal Pod Autoscaler (HPA)

## Advanced Topics

### Database Backups

Set up automated backups using CronJobs:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: supabase
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15
            command:
            - /bin/sh
            - -c
            - pg_dump -h supabase-db -U postgres > /backup/backup-$(date +%Y%m%d-%H%M%S).sql
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: postgres-backup-pvc
          restartPolicy: OnFailure
```

### Monitoring

Integrate with Prometheus and Grafana:

```yaml
# ServiceMonitor for Kong
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: supabase-kong
  namespace: supabase
spec:
  selector:
    matchLabels:
      app: supabase-kong
  endpoints:
  - port: metrics
    interval: 30s
```

### High Availability

For production:

1. **Multiple replicas**: Scale stateless services (Kong, Auth, etc.)
2. **Pod Disruption Budgets**: Prevent too many pods from being down
3. **Affinity rules**: Spread pods across nodes
4. **Database replication**: Consider PostgreSQL streaming replication

## Troubleshooting

See [troubleshooting.md](./troubleshooting.md) for common issues and solutions.

## Next Steps

- [Helm Installation Guide](./helm-installation.md)
- [Kustomize Installation Guide](./kustomize-installation.md)
- [Secret Management](./secret-extraction.md)
- [Architecture Overview](./architecture.md)

package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/converter"
	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/parser"
)

// Generator generates Helm charts
type Generator struct {
	resources *converter.K8sResources
	config    *parser.DockerCompose
	outputDir string
}

// NewGenerator creates a new Helm generator
func NewGenerator(resources *converter.K8sResources, config *parser.DockerCompose, outputDir string) *Generator {
	return &Generator{
		resources: resources,
		config:    config,
		outputDir: outputDir,
	}
}

// Generate generates the Helm chart
func (g *Generator) Generate() error {
	// Create directory structure
	dirs := []string{
		g.outputDir,
		filepath.Join(g.outputDir, "templates"),
		filepath.Join(g.outputDir, "templates", "postgres"),
		filepath.Join(g.outputDir, "templates", "kong"),
		filepath.Join(g.outputDir, "templates", "auth"),
		filepath.Join(g.outputDir, "templates", "rest"),
		filepath.Join(g.outputDir, "templates", "realtime"),
		filepath.Join(g.outputDir, "templates", "storage"),
		filepath.Join(g.outputDir, "templates", "meta"),
		filepath.Join(g.outputDir, "templates", "studio"),
		filepath.Join(g.outputDir, "templates", "functions"),
		filepath.Join(g.outputDir, "templates", "analytics"),
		filepath.Join(g.outputDir, "templates", "imgproxy"),
		filepath.Join(g.outputDir, "templates", "vector"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate Chart.yaml
	if err := g.generateChartYAML(); err != nil {
		return err
	}

	// Generate values.yaml
	if err := g.generateValuesYAML(); err != nil {
		return err
	}

	// Generate _helpers.tpl
	if err := g.generateHelpers(); err != nil {
		return err
	}

	// Generate NOTES.txt
	if err := g.generateNotes(); err != nil {
		return err
	}

	// Generate templates
	if err := g.generateTemplates(); err != nil {
		return err
	}

	return nil
}

// generateChartYAML generates Chart.yaml
func (g *Generator) generateChartYAML() error {
	chart := map[string]interface{}{
		"apiVersion": "v2",
		"name":       "supabase",
		"description": "A Helm chart for Supabase - The open source Firebase alternative",
		"type":       "application",
		"version":    "0.1.0",
		"appVersion": "latest",
		"keywords": []string{
			"supabase",
			"postgres",
			"database",
			"backend",
			"api",
		},
		"home": "https://supabase.com",
		"sources": []string{
			"https://github.com/supabase/supabase",
		},
		"maintainers": []map[string]string{
			{
				"name": "Supabase",
			},
		},
	}

	return g.writeYAML(filepath.Join(g.outputDir, "Chart.yaml"), chart)
}

// generateValuesYAML generates values.yaml
func (g *Generator) generateValuesYAML() error {
	values := map[string]interface{}{
		"global": map[string]interface{}{
			"namespace": "supabase",
		},
		"image": map[string]interface{}{
			"pullPolicy": "IfNotPresent",
		},
		"secrets": map[string]interface{}{
			"jwt": map[string]interface{}{
				"secret":       "your-super-secret-jwt-token-with-at-least-32-characters-long",
				"anonKey":      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0",
				"serviceRoleKey": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU",
			},
			"database": map[string]interface{}{
				"password": "your-super-secret-and-long-postgres-password",
			},
			"dashboard": map[string]interface{}{
				"username": "supabase",
				"password": "this_password_is_insecure_and_should_be_updated",
			},
		},
		"config": map[string]interface{}{
			"api": map[string]interface{}{
				"externalUrl": "http://localhost:8000",
			},
			"studio": map[string]interface{}{
				"port": 3000,
			},
		},
	}

	// Add service-specific values
	for name := range g.config.Services {
		serviceValues := g.generateServiceValues(name)
		if serviceValues != nil {
			values[name] = serviceValues
		}
	}

	return g.writeYAML(filepath.Join(g.outputDir, "values.yaml"), values)
}

// generateServiceValues generates values for a specific service
func (g *Generator) generateServiceValues(name string) map[string]interface{} {
	values := map[string]interface{}{
		"enabled":  true,
		"replicas": 1,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": "256Mi",
				"cpu":    "100m",
			},
			"limits": map[string]interface{}{
				"memory": "512Mi",
				"cpu":    "500m",
			},
		},
	}

	// Add service-specific overrides
	switch name {
	case "db":
		values["replicas"] = 1
		values["resources"] = map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": "1Gi",
				"cpu":    "500m",
			},
			"limits": map[string]interface{}{
				"memory": "2Gi",
				"cpu":    "1000m",
			},
		}
		values["persistence"] = map[string]interface{}{
			"enabled":      true,
			"storageClass": "",
			"size":         "10Gi",
		}
	case "kong":
		values["service"] = map[string]interface{}{
			"type": "LoadBalancer",
			"ports": map[string]interface{}{
				"http":  8000,
				"https": 8443,
			},
		}
	}

	return values
}

// generateHelpers generates _helpers.tpl
func (g *Generator) generateHelpers() error {
	helpers := `{{/*
Expand the name of the chart.
*/}}
{{- define "supabase.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "supabase.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "supabase.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "supabase.labels" -}}
helm.sh/chart: {{ include "supabase.chart" . }}
{{ include "supabase.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "supabase.selectorLabels" -}}
app.kubernetes.io/name: {{ include "supabase.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Component labels
*/}}
{{- define "supabase.componentLabels" -}}
{{ include "supabase.labels" . }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "supabase.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "supabase.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database URL
*/}}
{{- define "supabase.databaseUrl" -}}
postgresql://postgres:{{ .Values.secrets.database.password }}@{{ include "supabase.fullname" . }}-db:5432/postgres
{{- end }}

{{/*
API External URL
*/}}
{{- define "supabase.apiExternalUrl" -}}
{{- default (printf "http://%s-kong:8000" (include "supabase.fullname" .)) .Values.config.api.externalUrl }}
{{- end }}
`

	return os.WriteFile(filepath.Join(g.outputDir, "templates", "_helpers.tpl"), []byte(helpers), 0644)
}

// generateNotes generates NOTES.txt
func (g *Generator) generateNotes() error {
	notes := `Thank you for installing {{ .Chart.Name }}!

Your Supabase instance is being deployed.

To get the status of your deployment, run:
  kubectl get pods -n {{ .Values.global.namespace }}

To access the Supabase Studio dashboard:
  kubectl port-forward -n {{ .Values.global.namespace }} svc/{{ include "supabase.fullname" . }}-studio 3000:3000

Then visit http://localhost:3000 in your browser.

Dashboard Credentials:
  Username: {{ .Values.secrets.dashboard.username }}
  Password: {{ .Values.secrets.dashboard.password }}

API Gateway:
  kubectl port-forward -n {{ .Values.global.namespace }} svc/{{ include "supabase.fullname" . }}-kong 8000:8000

API Keys:
  Anon Key: {{ .Values.secrets.jwt.anonKey }}
  Service Role Key: {{ .Values.secrets.jwt.serviceRoleKey }}

IMPORTANT SECURITY NOTES:
  - Change all default passwords and secrets before using in production
  - Update JWT_SECRET in values.yaml
  - Configure proper ingress for external access
  - Enable TLS/SSL for production deployments
  - Review and adjust resource limits based on your workload

For more information, visit: https://supabase.com/docs
`

	return os.WriteFile(filepath.Join(g.outputDir, "templates", "NOTES.txt"), []byte(notes), 0644)
}

// generateTemplates generates all template files
func (g *Generator) generateTemplates() error {
	templatesDir := filepath.Join(g.outputDir, "templates")

	// Generate namespace template
	if err := g.generateNamespaceTemplate(templatesDir); err != nil {
		return err
	}

	// Generate ConfigMap templates
	for _, cm := range g.resources.ConfigMaps {
		if err := g.generateConfigMapTemplate(templatesDir, cm); err != nil {
			return err
		}
	}

	// Generate Secret templates
	for _, secret := range g.resources.Secrets {
		if err := g.generateSecretTemplate(templatesDir, secret); err != nil {
			return err
		}
	}

	// Generate Service, Deployment, and StatefulSet templates
	for name, svc := range g.resources.Services {
		if err := g.generateServiceTemplate(templatesDir, name, svc); err != nil {
			return err
		}
	}

	for name, deploy := range g.resources.Deployments {
		if err := g.generateDeploymentTemplate(templatesDir, name, deploy); err != nil {
			return err
		}
	}

	for name, ss := range g.resources.StatefulSets {
		if err := g.generateStatefulSetTemplate(templatesDir, name, ss); err != nil {
			return err
		}
	}

	// Generate ServiceAccount, Job, and RBAC templates
	if g.resources.ServiceAccount != nil {
		if err := g.generateServiceAccountTemplate(templatesDir); err != nil {
			return err
		}
	}

	if g.resources.SecretExtractorJob != nil {
		if err := g.generateJobTemplate(templatesDir); err != nil {
			return err
		}
	}

	if g.resources.SecretExtractorRBAC != nil {
		if err := g.generateRBACTemplate(templatesDir); err != nil {
			return err
		}
	}

	return nil
}

// generateNamespaceTemplate generates namespace template
func (g *Generator) generateNamespaceTemplate(templatesDir string) error {
	tmpl := `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
`
	return os.WriteFile(filepath.Join(templatesDir, "namespace.yaml"), []byte(tmpl), 0644)
}

// generateConfigMapTemplate generates ConfigMap template
func (g *Generator) generateConfigMapTemplate(templatesDir string, cm *converter.ConfigMap) error {
	tmplStr := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "supabase.fullname" . }}-%s
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
data:
`, cm.Name)

	for key := range cm.Data {
		tmplStr += fmt.Sprintf("  %s: {{ .Values.config.%s | quote }}\n", key, g.getValuesPath(key))
	}

	return os.WriteFile(filepath.Join(templatesDir, fmt.Sprintf("configmap-%s.yaml", cm.Name)), []byte(tmplStr), 0644)
}

// generateSecretTemplate generates Secret template
func (g *Generator) generateSecretTemplate(templatesDir string, secret *converter.Secret) error {
	tmplStr := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: {{ include "supabase.fullname" . }}-%s
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
type: %s
stringData:
`, secret.Name, secret.Type)

	for key := range secret.StringData {
		valuesPath := g.getSecretValuesPath(key)
		tmplStr += fmt.Sprintf("  %s: {{ .Values.%s | quote }}\n", key, valuesPath)
	}

	return os.WriteFile(filepath.Join(templatesDir, fmt.Sprintf("secret-%s.yaml", secret.Name)), []byte(tmplStr), 0644)
}

// generateServiceTemplate generates Service template
func (g *Generator) generateServiceTemplate(templatesDir string, name string, svc *converter.ServiceResource) error {
	serviceName := g.extractServiceName(svc.Name)

	tmplStr := fmt.Sprintf(`{{- if .Values.%s.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "supabase.fullname" . }}-%s
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
    app.kubernetes.io/component: %s
spec:
  type: {{ .Values.%s.service.type | default "ClusterIP" }}
  selector:
    app.kubernetes.io/name: {{ include "supabase.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/component: %s
  ports:
`, serviceName, serviceName, serviceName, serviceName, serviceName)

	for _, port := range svc.Ports {
		tmplStr += fmt.Sprintf(`    - name: %s
      port: %d
      targetPort: %d
      protocol: %s
`, port.Name, port.Port, port.TargetPort, port.Protocol)
	}

	tmplStr += "{{- end }}\n"

	// Determine subdirectory
	subdir := g.getServiceSubdir(serviceName)
	targetDir := filepath.Join(templatesDir, subdir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(targetDir, "service.yaml"), []byte(tmplStr), 0644)
}

// generateDeploymentTemplate generates Deployment template
func (g *Generator) generateDeploymentTemplate(templatesDir string, name string, deploy *converter.Deployment) error {
	serviceName := g.extractServiceName(deploy.Name)

	tmpl := `{{- if .Values.` + serviceName + `.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "supabase.fullname" . }}-` + serviceName + `
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
    app.kubernetes.io/component: ` + serviceName + `
spec:
  replicas: {{ .Values.` + serviceName + `.replicas }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "supabase.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
      app.kubernetes.io/component: ` + serviceName + `
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ include "supabase.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
        app.kubernetes.io/component: ` + serviceName + `
    spec:
      containers:`

	// Add container spec
	for _, container := range deploy.Template.Containers {
		tmpl += fmt.Sprintf(`
        - name: %s
          image: %s
          imagePullPolicy: {{ .Values.image.pullPolicy }}`, container.Name, container.Image)

		if len(container.Ports) > 0 {
			tmpl += "\n          ports:"
			for _, port := range container.Ports {
				tmpl += fmt.Sprintf(`
            - name: %s
              containerPort: %d
              protocol: %s`, port.Name, port.ContainerPort, port.Protocol)
			}
		}

		if len(container.Env) > 0 {
			tmpl += "\n          env:"
			for _, env := range container.Env {
				if env.Value != "" {
					tmpl += fmt.Sprintf(`
            - name: %s
              value: %s`, env.Name, env.Value)
				} else if env.ValueFrom != nil {
					tmpl += fmt.Sprintf(`
            - name: %s
              valueFrom:`, env.Name)
					if env.ValueFrom.SecretKeyRef != nil {
						tmpl += fmt.Sprintf(`
                secretKeyRef:
                  name: {{ include "supabase.fullname" . }}-%s
                  key: %s`, env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
					} else if env.ValueFrom.ConfigMapKeyRef != nil {
						tmpl += fmt.Sprintf(`
                configMapKeyRef:
                  name: {{ include "supabase.fullname" . }}-%s
                  key: %s`, env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key)
					}
				}
			}
		}

		tmpl += `
          resources:
            {{- toYaml .Values.` + serviceName + `.resources | nindent 12 }}`
	}

	tmpl += "\n{{- end }}\n"

	subdir := g.getServiceSubdir(serviceName)
	targetDir := filepath.Join(templatesDir, subdir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(targetDir, "deployment.yaml"), []byte(tmpl), 0644)
}

// generateStatefulSetTemplate generates StatefulSet template
func (g *Generator) generateStatefulSetTemplate(templatesDir string, name string, ss *converter.StatefulSet) error {
	serviceName := g.extractServiceName(ss.Name)

	tmpl := `{{- if .Values.` + serviceName + `.enabled }}
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ include "supabase.fullname" . }}-` + serviceName + `
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
    app.kubernetes.io/component: ` + serviceName + `
spec:
  serviceName: {{ include "supabase.fullname" . }}-` + serviceName + `
  replicas: {{ .Values.` + serviceName + `.replicas }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "supabase.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
      app.kubernetes.io/component: ` + serviceName + `
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ include "supabase.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
        app.kubernetes.io/component: ` + serviceName + `
    spec:
      containers:`

	// Add container spec
	for _, container := range ss.Template.Containers {
		tmpl += fmt.Sprintf(`
        - name: %s
          image: %s
          imagePullPolicy: {{ .Values.image.pullPolicy }}`, container.Name, container.Image)

		if len(container.Ports) > 0 {
			tmpl += "\n          ports:"
			for _, port := range container.Ports {
				tmpl += fmt.Sprintf(`
            - name: %s
              containerPort: %d
              protocol: %s`, port.Name, port.ContainerPort, port.Protocol)
			}
		}

		if len(container.Env) > 0 {
			tmpl += "\n          env:"
			for _, env := range container.Env {
				if env.Value != "" {
					tmpl += fmt.Sprintf(`
            - name: %s
              value: "%s"`, env.Name, env.Value)
				}
			}
		}

		tmpl += `
          resources:
            {{- toYaml .Values.` + serviceName + `.resources | nindent 12 }}`

		if len(container.VolumeMounts) > 0 {
			tmpl += "\n          volumeMounts:"
			for _, mount := range container.VolumeMounts {
				tmpl += fmt.Sprintf(`
            - name: %s
              mountPath: %s`, mount.Name, mount.MountPath)
			}
		}
	}

	// Add volume claim templates
	if len(ss.VolumeClaimTemplates) > 0 {
		tmpl += `
  volumeClaimTemplates:`
		for _, vct := range ss.VolumeClaimTemplates {
			tmpl += fmt.Sprintf(`
    - metadata:
        name: %s
      spec:
        accessModes:`, vct.Name)
			for _, mode := range vct.AccessModes {
				tmpl += fmt.Sprintf(`
          - %s`, mode)
			}
			tmpl += fmt.Sprintf(`
        {{- if .Values.%s.persistence.storageClass }}
        storageClassName: {{ .Values.%s.persistence.storageClass }}
        {{- end }}
        resources:
          requests:
            storage: {{ .Values.%s.persistence.size }}`, serviceName, serviceName, serviceName)
		}
	}

	tmpl += "\n{{- end }}\n"

	subdir := g.getServiceSubdir(serviceName)
	targetDir := filepath.Join(templatesDir, subdir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(targetDir, "statefulset.yaml"), []byte(tmpl), 0644)
}

// generateServiceAccountTemplate generates ServiceAccount template
func (g *Generator) generateServiceAccountTemplate(templatesDir string) error {
	tmpl := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "supabase.fullname" . }}-secret-extractor
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
`
	return os.WriteFile(filepath.Join(templatesDir, "serviceaccount.yaml"), []byte(tmpl), 0644)
}

// generateJobTemplate generates Job template
func (g *Generator) generateJobTemplate(templatesDir string) error {
	tmpl := `apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "supabase.fullname" . }}-secret-extractor
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "10"
    "helm.sh/hook-delete-policy": hook-succeeded
spec:
  template:
    metadata:
      labels:
        app: supabase-secret-extractor
    spec:
      serviceAccountName: {{ include "supabase.fullname" . }}-secret-extractor
      restartPolicy: Never
      containers:
        - name: secret-extractor
          image: bitnami/kubectl:latest
          command: ["/bin/sh", "-c"]
          args:
            - |
              echo "================================================================"
              echo "Supabase Secrets"
              echo "================================================================"
              echo ""
              echo "Dashboard Credentials:"
              echo "  Username: $(kubectl get secret {{ include "supabase.fullname" . }}-supabase-secrets -n {{ .Values.global.namespace }} -o jsonpath='{.data.DASHBOARD_USERNAME}' | base64 -d)"
              echo "  Password: $(kubectl get secret {{ include "supabase.fullname" . }}-supabase-secrets -n {{ .Values.global.namespace }} -o jsonpath='{.data.DASHBOARD_PASSWORD}' | base64 -d)"
              echo ""
              echo "JWT Tokens:"
              echo "  Anon Key: $(kubectl get secret {{ include "supabase.fullname" . }}-supabase-jwt -n {{ .Values.global.namespace }} -o jsonpath='{.data.ANON_KEY}' | base64 -d)"
              echo "  Service Role Key: $(kubectl get secret {{ include "supabase.fullname" . }}-supabase-jwt -n {{ .Values.global.namespace }} -o jsonpath='{.data.SERVICE_ROLE_KEY}' | base64 -d)"
              echo ""
              echo "Database:"
              echo "  Password: $(kubectl get secret {{ include "supabase.fullname" . }}-supabase-db -n {{ .Values.global.namespace }} -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)"
              echo ""
              echo "================================================================"
`
	return os.WriteFile(filepath.Join(templatesDir, "secret-extractor-job.yaml"), []byte(tmpl), 0644)
}

// generateRBACTemplate generates RBAC template
func (g *Generator) generateRBACTemplate(templatesDir string) error {
	tmpl := `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ include "supabase.fullname" . }}-secret-reader
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ include "supabase.fullname" . }}-secret-reader-binding
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "supabase.labels" . | nindent 4 }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ include "supabase.fullname" . }}-secret-reader
subjects:
  - kind: ServiceAccount
    name: {{ include "supabase.fullname" . }}-secret-extractor
    namespace: {{ .Values.global.namespace }}
`
	return os.WriteFile(filepath.Join(templatesDir, "secret-extractor-rbac.yaml"), []byte(tmpl), 0644)
}

// Helper functions

func (g *Generator) extractServiceName(fullName string) string {
	// Remove "supabase-" prefix
	return strings.TrimPrefix(fullName, "supabase-")
}

func (g *Generator) getServiceSubdir(serviceName string) string {
	// Map service names to subdirectories
	subdirs := map[string]string{
		"db":        "postgres",
		"kong":      "kong",
		"auth":      "auth",
		"rest":      "rest",
		"realtime":  "realtime",
		"storage":   "storage",
		"meta":      "meta",
		"studio":    "studio",
		"functions": "functions",
		"analytics": "analytics",
		"imgproxy":  "imgproxy",
		"vector":    "vector",
	}

	if subdir, ok := subdirs[serviceName]; ok {
		return subdir
	}
	return ""
}

func (g *Generator) getValuesPath(key string) string {
	// Simple mapping for demo purposes
	// In production, this would be more sophisticated
	return strings.ToLower(strings.ReplaceAll(key, "_", "."))
}

func (g *Generator) getSecretValuesPath(key string) string {
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "jwt") || strings.Contains(keyLower, "anon") || strings.Contains(keyLower, "service_role") {
		return "secrets.jwt." + g.camelCase(key)
	}
	if strings.Contains(keyLower, "postgres") || strings.Contains(keyLower, "db") {
		return "secrets.database." + g.camelCase(key)
	}
	if strings.Contains(keyLower, "dashboard") {
		return "secrets.dashboard." + g.camelCase(key)
	}
	return "secrets." + g.camelCase(key)
}

func (g *Generator) camelCase(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func (g *Generator) writeYAML(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

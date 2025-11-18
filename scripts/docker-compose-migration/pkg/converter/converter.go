package converter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/parser"
)

// K8sResources holds all converted Kubernetes resources
type K8sResources struct {
	Namespace      *Namespace
	ConfigMaps     map[string]*ConfigMap
	Secrets        map[string]*Secret
	Services       map[string]*ServiceResource
	Deployments    map[string]*Deployment
	StatefulSets   map[string]*StatefulSet
	PVCs           map[string]*PVC
	ServiceAccount *ServiceAccount
	SecretExtractorJob *Job
	SecretExtractorRBAC *RBAC
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name string
}

// ConfigMap represents a Kubernetes ConfigMap
type ConfigMap struct {
	Name string
	Data map[string]string
}

// Secret represents a Kubernetes Secret
type Secret struct {
	Name        string
	Data        map[string]string
	StringData  map[string]string
	Type        string
}

// ServiceResource represents a Kubernetes Service
type ServiceResource struct {
	Name      string
	Selector  map[string]string
	Ports     []ServicePort
	Type      string
	Labels    map[string]string
}

// ServicePort represents a Kubernetes Service port
type ServicePort struct {
	Name       string
	Port       int32
	TargetPort int32
	Protocol   string
}

// Deployment represents a Kubernetes Deployment
type Deployment struct {
	Name     string
	Replicas int32
	Labels   map[string]string
	Selector map[string]string
	Template PodTemplate
}

// StatefulSet represents a Kubernetes StatefulSet
type StatefulSet struct {
	Name             string
	Replicas         int32
	ServiceName      string
	Labels           map[string]string
	Selector         map[string]string
	Template         PodTemplate
	VolumeClaimTemplates []PVCTemplate
}

// PodTemplate represents a Kubernetes Pod template
type PodTemplate struct {
	Labels      map[string]string
	Annotations map[string]string
	Containers  []Container
	Volumes     []Volume
	InitContainers []Container
}

// Container represents a Kubernetes container
type Container struct {
	Name            string
	Image           string
	Command         []string
	Args            []string
	Env             []EnvVar
	Ports           []ContainerPort
	VolumeMounts    []VolumeMount
	LivenessProbe   *Probe
	ReadinessProbe  *Probe
	Resources       *Resources
	SecurityContext *SecurityContext
}

// EnvVar represents a Kubernetes environment variable
type EnvVar struct {
	Name      string
	Value     string
	ValueFrom *EnvVarSource
}

// EnvVarSource represents the source of an environment variable
type EnvVarSource struct {
	SecretKeyRef    *KeyRef
	ConfigMapKeyRef *KeyRef
	FieldRef        *FieldRef
}

// KeyRef represents a reference to a key in a Secret or ConfigMap
type KeyRef struct {
	Name string
	Key  string
}

// FieldRef represents a reference to a field
type FieldRef struct {
	FieldPath string
}

// ContainerPort represents a Kubernetes container port
type ContainerPort struct {
	Name          string
	ContainerPort int32
	Protocol      string
}

// VolumeMount represents a Kubernetes volume mount
type VolumeMount struct {
	Name      string
	MountPath string
	SubPath   string
	ReadOnly  bool
}

// Volume represents a Kubernetes volume
type Volume struct {
	Name         string
	VolumeSource VolumeSource
}

// VolumeSource represents a Kubernetes volume source
type VolumeSource struct {
	ConfigMap              *ConfigMapVolumeSource
	Secret                 *SecretVolumeSource
	PersistentVolumeClaim  *PVCVolumeSource
	EmptyDir               *EmptyDirVolumeSource
	HostPath               *HostPathVolumeSource
}

// ConfigMapVolumeSource represents a ConfigMap volume source
type ConfigMapVolumeSource struct {
	Name string
}

// SecretVolumeSource represents a Secret volume source
type SecretVolumeSource struct {
	SecretName string
}

// PVCVolumeSource represents a PVC volume source
type PVCVolumeSource struct {
	ClaimName string
}

// EmptyDirVolumeSource represents an emptyDir volume source
type EmptyDirVolumeSource struct {
	Medium string
}

// HostPathVolumeSource represents a hostPath volume source
type HostPathVolumeSource struct {
	Path string
	Type string
}

// Probe represents a Kubernetes probe
type Probe struct {
	Exec                *ExecAction
	HTTPGet             *HTTPGetAction
	TCPSocket           *TCPSocketAction
	InitialDelaySeconds int32
	PeriodSeconds       int32
	TimeoutSeconds      int32
	SuccessThreshold    int32
	FailureThreshold    int32
}

// ExecAction represents a Kubernetes exec action
type ExecAction struct {
	Command []string
}

// HTTPGetAction represents a Kubernetes HTTP GET action
type HTTPGetAction struct {
	Path   string
	Port   int32
	Scheme string
}

// TCPSocketAction represents a Kubernetes TCP socket action
type TCPSocketAction struct {
	Port int32
}

// Resources represents Kubernetes resource requirements
type Resources struct {
	Requests map[string]string
	Limits   map[string]string
}

// SecurityContext represents a Kubernetes security context
type SecurityContext struct {
	Capabilities   *Capabilities
	Privileged     *bool
	RunAsUser      *int64
	RunAsNonRoot   *bool
	ReadOnlyRootFilesystem *bool
}

// Capabilities represents Linux capabilities
type Capabilities struct {
	Add  []string
	Drop []string
}

// PVC represents a PersistentVolumeClaim
type PVC struct {
	Name         string
	StorageClass string
	AccessModes  []string
	Size         string
}

// PVCTemplate represents a PVC template for StatefulSet
type PVCTemplate struct {
	Name         string
	StorageClass string
	AccessModes  []string
	Size         string
}

// ServiceAccount represents a Kubernetes ServiceAccount
type ServiceAccount struct {
	Name string
}

// Job represents a Kubernetes Job
type Job struct {
	Name     string
	Labels   map[string]string
	Template PodTemplate
	Annotations map[string]string
}

// RBAC represents Kubernetes RBAC resources
type RBAC struct {
	Role        *Role
	RoleBinding *RoleBinding
}

// Role represents a Kubernetes Role
type Role struct {
	Name  string
	Rules []PolicyRule
}

// PolicyRule represents a Kubernetes policy rule
type PolicyRule struct {
	APIGroups []string
	Resources []string
	Verbs     []string
}

// RoleBinding represents a Kubernetes RoleBinding
type RoleBinding struct {
	Name               string
	RoleRef            RoleRef
	Subjects           []Subject
}

// RoleRef represents a role reference
type RoleRef struct {
	APIGroup string
	Kind     string
	Name     string
}

// Subject represents a subject in a role binding
type Subject struct {
	Kind      string
	Name      string
	Namespace string
}

// Converter converts docker-compose to Kubernetes resources
type Converter struct {
	config *parser.DockerCompose
}

// NewConverter creates a new Converter instance
func NewConverter(config *parser.DockerCompose) *Converter {
	return &Converter{config: config}
}

// Convert converts docker-compose configuration to Kubernetes resources
func (c *Converter) Convert() (*K8sResources, error) {
	resources := &K8sResources{
		Namespace: &Namespace{
			Name: "supabase",
		},
		ConfigMaps:   make(map[string]*ConfigMap),
		Secrets:      make(map[string]*Secret),
		Services:     make(map[string]*ServiceResource),
		Deployments:  make(map[string]*Deployment),
		StatefulSets: make(map[string]*StatefulSet),
		PVCs:         make(map[string]*PVC),
	}

	// Create common ConfigMap for shared configuration
	resources.ConfigMaps["supabase-config"] = c.createCommonConfigMap()

	// Create Secrets
	resources.Secrets["supabase-secrets"] = c.createSecretsResource()
	resources.Secrets["supabase-jwt"] = c.createJWTSecret()
	resources.Secrets["supabase-db"] = c.createDBSecret()

	// Convert each service
	for name, service := range c.config.Services {
		if err := c.convertService(name, service, resources); err != nil {
			return nil, fmt.Errorf("failed to convert service %s: %w", name, err)
		}
	}

	// Create secret extractor job and RBAC
	resources.ServiceAccount = &ServiceAccount{Name: "supabase-secret-extractor"}
	resources.SecretExtractorJob = c.createSecretExtractorJob()
	resources.SecretExtractorRBAC = c.createSecretExtractorRBAC()

	return resources, nil
}

// createCommonConfigMap creates a common ConfigMap for shared configuration
func (c *Converter) createCommonConfigMap() *ConfigMap {
	return &ConfigMap{
		Name: "supabase-config",
		Data: map[string]string{
			"POSTGRES_HOST":     "supabase-db",
			"POSTGRES_PORT":     "5432",
			"POSTGRES_DB":       "postgres",
			"KONG_HTTP_PORT":    "8000",
			"KONG_HTTPS_PORT":   "8443",
			"API_EXTERNAL_URL":  "http://kong:8000",
			"PGRST_DB_SCHEMAS":  "public,storage,graphql_public",
			"ENABLE_EMAIL_SIGNUP": "true",
			"ENABLE_EMAIL_AUTOCONFIRM": "false",
		},
	}
}

// createSecretsResource creates the main secrets resource
func (c *Converter) createSecretsResource() *Secret {
	return &Secret{
		Name: "supabase-secrets",
		Type: "Opaque",
		StringData: map[string]string{
			"DASHBOARD_USERNAME": "supabase",
			"DASHBOARD_PASSWORD": "this_password_is_insecure_and_should_be_updated",
			"SECRET_KEY_BASE":    "UpNVntn3cDxHJpq99YMc1T1AQgQpc8kfYTuRgBiYa15BLrx8etQoXz3gZv1/u2oq",
		},
	}
}

// createJWTSecret creates JWT secret
func (c *Converter) createJWTSecret() *Secret {
	return &Secret{
		Name: "supabase-jwt",
		Type: "Opaque",
		StringData: map[string]string{
			"JWT_SECRET":        "your-super-secret-jwt-token-with-at-least-32-characters-long",
			"ANON_KEY":          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0",
			"SERVICE_ROLE_KEY":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU",
		},
	}
}

// createDBSecret creates database secret
func (c *Converter) createDBSecret() *Secret {
	return &Secret{
		Name: "supabase-db",
		Type: "Opaque",
		StringData: map[string]string{
			"POSTGRES_PASSWORD": "your-super-secret-and-long-postgres-password",
		},
	}
}

// convertService converts a docker-compose service to Kubernetes resources
func (c *Converter) convertService(name string, service *parser.Service, resources *K8sResources) error {
	// Determine if service should be a StatefulSet (database, storage)
	isStateful := c.isStatefulService(name)

	// Create Service resource if ports are exposed
	if len(service.Ports) > 0 {
		resources.Services[name] = c.createServiceResource(name, service)
	}

	// Create Deployment or StatefulSet
	if isStateful {
		resources.StatefulSets[name] = c.createStatefulSet(name, service)
	} else {
		resources.Deployments[name] = c.createDeployment(name, service)
	}

	return nil
}

// isStatefulService determines if a service should be a StatefulSet
func (c *Converter) isStatefulService(name string) bool {
	statefulServices := []string{"db", "storage"}
	for _, s := range statefulServices {
		if strings.Contains(name, s) {
			return true
		}
	}
	return false
}

// createServiceResource creates a Kubernetes Service
func (c *Converter) createServiceResource(name string, service *parser.Service) *ServiceResource {
	svc := &ServiceResource{
		Name:     fmt.Sprintf("supabase-%s", name),
		Selector: map[string]string{"app": fmt.Sprintf("supabase-%s", name)},
		Type:     "ClusterIP",
		Labels:   map[string]string{"app": fmt.Sprintf("supabase-%s", name)},
		Ports:    []ServicePort{},
	}

	// Parse ports
	for _, portMapping := range service.Ports {
		port := c.parsePort(portMapping)
		if port != nil {
			svc.Ports = append(svc.Ports, *port)
		}
	}

	return svc
}

// parsePort parses a docker-compose port mapping
func (c *Converter) parsePort(portMapping string) *ServicePort {
	// Format: "host:container" or "container"
	parts := strings.Split(portMapping, ":")
	var containerPort int
	var hostPort int

	if len(parts) == 2 {
		hostPort, _ = strconv.Atoi(parts[0])
		containerPort, _ = strconv.Atoi(parts[1])
	} else if len(parts) == 1 {
		containerPort, _ = strconv.Atoi(parts[0])
		hostPort = containerPort
	} else {
		return nil
	}

	return &ServicePort{
		Name:       fmt.Sprintf("port-%d", containerPort),
		Port:       int32(hostPort),
		TargetPort: int32(containerPort),
		Protocol:   "TCP",
	}
}

// createDeployment creates a Kubernetes Deployment
func (c *Converter) createDeployment(name string, service *parser.Service) *Deployment {
	labels := map[string]string{"app": fmt.Sprintf("supabase-%s", name)}

	return &Deployment{
		Name:     fmt.Sprintf("supabase-%s", name),
		Replicas: 1,
		Labels:   labels,
		Selector: labels,
		Template: c.createPodTemplate(name, service, labels),
	}
}

// createStatefulSet creates a Kubernetes StatefulSet
func (c *Converter) createStatefulSet(name string, service *parser.Service) *StatefulSet {
	labels := map[string]string{"app": fmt.Sprintf("supabase-%s", name)}

	ss := &StatefulSet{
		Name:        fmt.Sprintf("supabase-%s", name),
		Replicas:    1,
		ServiceName: fmt.Sprintf("supabase-%s", name),
		Labels:      labels,
		Selector:    labels,
		Template:    c.createPodTemplate(name, service, labels),
		VolumeClaimTemplates: []PVCTemplate{},
	}

	// Add PVC template for stateful services
	if name == "db" {
		ss.VolumeClaimTemplates = append(ss.VolumeClaimTemplates, PVCTemplate{
			Name:         "postgres-data",
			AccessModes:  []string{"ReadWriteOnce"},
			Size:         "10Gi",
			StorageClass: "",
		})
	}

	return ss
}

// createPodTemplate creates a Pod template
func (c *Converter) createPodTemplate(name string, service *parser.Service, labels map[string]string) PodTemplate {
	container := c.createContainer(name, service)

	return PodTemplate{
		Labels:     labels,
		Containers: []Container{container},
		Volumes:    c.createVolumes(name, service),
	}
}

// createContainer creates a container spec
func (c *Converter) createContainer(name string, service *parser.Service) Container {
	container := Container{
		Name:         name,
		Image:        service.Image,
		Command:      service.GetCommandAsSlice(),
		Env:          c.createEnvVars(name, service),
		Ports:        c.createContainerPorts(service),
		VolumeMounts: c.createVolumeMounts(name, service),
	}

	// Add health checks
	if service.HealthCheck != nil {
		probe := c.convertHealthCheck(service.HealthCheck)
		container.LivenessProbe = probe
		container.ReadinessProbe = probe
	}

	// Add security context
	if service.Privileged || len(service.CapAdd) > 0 || len(service.CapDrop) > 0 {
		container.SecurityContext = c.createSecurityContext(service)
	}

	return container
}

// createEnvVars creates environment variables for a container
func (c *Converter) createEnvVars(name string, service *parser.Service) []EnvVar {
	envVars := []EnvVar{}

	// Process environment map
	for key, value := range service.Environment {
		envVar := c.createEnvVar(key, value)
		if envVar != nil {
			envVars = append(envVars, *envVar)
		}
	}

	return envVars
}

// createEnvVar creates a single environment variable
func (c *Converter) createEnvVar(key string, value interface{}) *EnvVar {
	// Determine if this should come from a secret or configmap
	if c.isSecretEnvVar(key) {
		return &EnvVar{
			Name: key,
			ValueFrom: &EnvVarSource{
				SecretKeyRef: &KeyRef{
					Name: c.getSecretNameForKey(key),
					Key:  key,
				},
			},
		}
	}

	// Check if it's a config value
	if c.isConfigEnvVar(key) {
		return &EnvVar{
			Name: key,
			ValueFrom: &EnvVarSource{
				ConfigMapKeyRef: &KeyRef{
					Name: "supabase-config",
					Key:  key,
				},
			},
		}
	}

	// Regular value
	valueStr := fmt.Sprintf("%v", value)
	return &EnvVar{
		Name:  key,
		Value: valueStr,
	}
}

// isSecretEnvVar determines if an env var should come from a secret
func (c *Converter) isSecretEnvVar(key string) bool {
	secretKeys := []string{
		"PASSWORD", "SECRET", "KEY", "TOKEN", "CREDENTIAL",
	}
	keyUpper := strings.ToUpper(key)
	for _, secretKey := range secretKeys {
		if strings.Contains(keyUpper, secretKey) {
			return true
		}
	}
	return false
}

// isConfigEnvVar determines if an env var should come from a configmap
func (c *Converter) isConfigEnvVar(key string) bool {
	configKeys := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB",
		"KONG_HTTP_PORT", "KONG_HTTPS_PORT", "API_EXTERNAL_URL",
		"PGRST_DB_SCHEMAS", "ENABLE_EMAIL_SIGNUP", "ENABLE_EMAIL_AUTOCONFIRM",
	}
	for _, configKey := range configKeys {
		if key == configKey {
			return true
		}
	}
	return false
}

// getSecretNameForKey returns the appropriate secret name for a key
func (c *Converter) getSecretNameForKey(key string) string {
	keyUpper := strings.ToUpper(key)
	if strings.Contains(keyUpper, "JWT") || strings.Contains(keyUpper, "ANON") || strings.Contains(keyUpper, "SERVICE_ROLE") {
		return "supabase-jwt"
	}
	if strings.Contains(keyUpper, "POSTGRES") || strings.Contains(keyUpper, "DB") {
		return "supabase-db"
	}
	return "supabase-secrets"
}

// createContainerPorts creates container ports
func (c *Converter) createContainerPorts(service *parser.Service) []ContainerPort {
	ports := []ContainerPort{}
	for _, portMapping := range service.Ports {
		port := c.parsePort(portMapping)
		if port != nil {
			ports = append(ports, ContainerPort{
				Name:          port.Name,
				ContainerPort: port.TargetPort,
				Protocol:      port.Protocol,
			})
		}
	}
	return ports
}

// createVolumeMounts creates volume mounts
func (c *Converter) createVolumeMounts(name string, service *parser.Service) []VolumeMount {
	mounts := []VolumeMount{}
	for _, volumeMapping := range service.Volumes {
		mount := c.parseVolumeMount(volumeMapping)
		if mount != nil {
			mounts = append(mounts, *mount)
		}
	}
	return mounts
}

// parseVolumeMount parses a docker-compose volume mapping
func (c *Converter) parseVolumeMount(volumeMapping string) *VolumeMount {
	// Format: "host:container" or "named:container" or "container"
	parts := strings.Split(volumeMapping, ":")
	if len(parts) < 2 {
		return nil
	}

	source := parts[0]
	target := parts[1]
	readOnly := false

	if len(parts) == 3 && parts[2] == "ro" {
		readOnly = true
	}

	// Generate a volume name
	volumeName := c.sanitizeName(source)

	return &VolumeMount{
		Name:      volumeName,
		MountPath: target,
		ReadOnly:  readOnly,
	}
}

// createVolumes creates volumes
func (c *Converter) createVolumes(name string, service *parser.Service) []Volume {
	volumes := []Volume{}
	volumeMap := make(map[string]bool)

	for _, volumeMapping := range service.Volumes {
		parts := strings.Split(volumeMapping, ":")
		if len(parts) < 2 {
			continue
		}

		source := parts[0]
		volumeName := c.sanitizeName(source)

		// Avoid duplicates
		if volumeMap[volumeName] {
			continue
		}
		volumeMap[volumeName] = true

		// Determine volume source type
		var volumeSource VolumeSource

		// Check if it's a named volume (exists in docker-compose volumes)
		if _, exists := c.config.Volumes[source]; exists {
			volumeSource = VolumeSource{
				PersistentVolumeClaim: &PVCVolumeSource{
					ClaimName: volumeName,
				},
			}
		} else if strings.HasPrefix(source, "/") || strings.HasPrefix(source, "./") {
			// Host path
			volumeSource = VolumeSource{
				HostPath: &HostPathVolumeSource{
					Path: source,
					Type: "DirectoryOrCreate",
				},
			}
		} else {
			// EmptyDir
			volumeSource = VolumeSource{
				EmptyDir: &EmptyDirVolumeSource{},
			}
		}

		volumes = append(volumes, Volume{
			Name:         volumeName,
			VolumeSource: volumeSource,
		})
	}

	return volumes
}

// convertHealthCheck converts a docker-compose health check to Kubernetes probe
func (c *Converter) convertHealthCheck(hc *parser.HealthCheck) *Probe {
	probe := &Probe{
		TimeoutSeconds:   5,
		PeriodSeconds:    10,
		FailureThreshold: 3,
		SuccessThreshold: 1,
	}

	if hc.Timeout != "" {
		probe.TimeoutSeconds = c.parseDuration(hc.Timeout)
	}
	if hc.Interval != "" {
		probe.PeriodSeconds = c.parseDuration(hc.Interval)
	}
	if hc.Retries > 0 {
		probe.FailureThreshold = int32(hc.Retries)
	}

	// Parse test command
	var testCmd []string
	switch v := hc.Test.(type) {
	case []interface{}:
		for _, cmd := range v {
			testCmd = append(testCmd, fmt.Sprintf("%v", cmd))
		}
	case string:
		testCmd = []string{"/bin/sh", "-c", v}
	}

	if len(testCmd) > 0 {
		// Try to convert to HTTP or TCP probe if possible
		if c.isHTTPProbe(testCmd) {
			probe.HTTPGet = c.parseHTTPProbe(testCmd)
		} else if c.isTCPProbe(testCmd) {
			probe.TCPSocket = c.parseTCPProbe(testCmd)
		} else {
			// Skip CMD or CMD-SHELL prefix
			if len(testCmd) > 0 && (testCmd[0] == "CMD" || testCmd[0] == "CMD-SHELL") {
				testCmd = testCmd[1:]
			}
			probe.Exec = &ExecAction{Command: testCmd}
		}
	}

	return probe
}

// isHTTPProbe checks if a health check is an HTTP probe
func (c *Converter) isHTTPProbe(cmd []string) bool {
	cmdStr := strings.Join(cmd, " ")
	return strings.Contains(cmdStr, "curl") || strings.Contains(cmdStr, "wget") || strings.Contains(cmdStr, "http://") || strings.Contains(cmdStr, "https://")
}

// isTCPProbe checks if a health check is a TCP probe
func (c *Converter) isTCPProbe(cmd []string) bool {
	cmdStr := strings.Join(cmd, " ")
	return strings.Contains(cmdStr, "nc ") || strings.Contains(cmdStr, "netcat") || strings.Contains(cmdStr, "telnet")
}

// parseHTTPProbe parses an HTTP probe from a command
func (c *Converter) parseHTTPProbe(cmd []string) *HTTPGetAction {
	cmdStr := strings.Join(cmd, " ")

	// Extract URL
	re := regexp.MustCompile(`https?://[^\s]+`)
	matches := re.FindStringSubmatch(cmdStr)

	if len(matches) > 0 {
		url := matches[0]
		scheme := "http"
		if strings.HasPrefix(url, "https") {
			scheme = "https"
		}

		// Extract path and port
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "https://")

		parts := strings.Split(url, "/")
		hostPort := parts[0]
		path := "/"
		if len(parts) > 1 {
			path = "/" + strings.Join(parts[1:], "/")
		}

		// Extract port
		portParts := strings.Split(hostPort, ":")
		port := 80
		if scheme == "https" {
			port = 443
		}
		if len(portParts) > 1 {
			port, _ = strconv.Atoi(portParts[1])
		}

		return &HTTPGetAction{
			Path:   path,
			Port:   int32(port),
			Scheme: strings.ToUpper(scheme),
		}
	}

	return &HTTPGetAction{
		Path:   "/",
		Port:   8000,
		Scheme: "HTTP",
	}
}

// parseTCPProbe parses a TCP probe from a command
func (c *Converter) parseTCPProbe(cmd []string) *TCPSocketAction {
	cmdStr := strings.Join(cmd, " ")

	// Try to extract port number
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(cmdStr, -1)

	port := 8000
	if len(matches) > 0 {
		port, _ = strconv.Atoi(matches[len(matches)-1])
	}

	return &TCPSocketAction{
		Port: int32(port),
	}
}

// parseDuration parses a docker-compose duration to seconds
func (c *Converter) parseDuration(duration string) int32 {
	// Simple parser for durations like "30s", "1m", "1h"
	duration = strings.TrimSpace(duration)
	if duration == "" {
		return 30
	}

	re := regexp.MustCompile(`(\d+)([smh])`)
	matches := re.FindStringSubmatch(duration)

	if len(matches) == 3 {
		value, _ := strconv.Atoi(matches[1])
		unit := matches[2]

		switch unit {
		case "s":
			return int32(value)
		case "m":
			return int32(value * 60)
		case "h":
			return int32(value * 3600)
		}
	}

	return 30
}

// createSecurityContext creates a security context
func (c *Converter) createSecurityContext(service *parser.Service) *SecurityContext {
	sc := &SecurityContext{}

	if service.Privileged {
		privileged := true
		sc.Privileged = &privileged
	}

	if len(service.CapAdd) > 0 || len(service.CapDrop) > 0 {
		sc.Capabilities = &Capabilities{
			Add:  service.CapAdd,
			Drop: service.CapDrop,
		}
	}

	return sc
}

// createSecretExtractorJob creates a Job for extracting and outputting secrets
func (c *Converter) createSecretExtractorJob() *Job {
	return &Job{
		Name: "supabase-secret-extractor",
		Labels: map[string]string{
			"app": "supabase-secret-extractor",
		},
		Annotations: map[string]string{
			"helm.sh/hook":               "post-install,post-upgrade",
			"helm.sh/hook-weight":        "10",
			"helm.sh/hook-delete-policy": "hook-succeeded",
		},
		Template: PodTemplate{
			Labels: map[string]string{
				"app": "supabase-secret-extractor",
			},
			Containers: []Container{
				{
					Name:  "secret-extractor",
					Image: "bitnami/kubectl:latest",
					Command: []string{"/bin/sh", "-c"},
					Args: []string{`
echo "================================================================"
echo "Supabase Secrets"
echo "================================================================"
echo ""
echo "Dashboard Credentials:"
echo "  Username: $(kubectl get secret supabase-secrets -o jsonpath='{.data.DASHBOARD_USERNAME}' | base64 -d)"
echo "  Password: $(kubectl get secret supabase-secrets -o jsonpath='{.data.DASHBOARD_PASSWORD}' | base64 -d)"
echo ""
echo "JWT Tokens:"
echo "  Anon Key: $(kubectl get secret supabase-jwt -o jsonpath='{.data.ANON_KEY}' | base64 -d)"
echo "  Service Role Key: $(kubectl get secret supabase-jwt -o jsonpath='{.data.SERVICE_ROLE_KEY}' | base64 -d)"
echo ""
echo "Database:"
echo "  Password: $(kubectl get secret supabase-db -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)"
echo ""
echo "================================================================"
					`},
				},
			},
		},
	}
}

// createSecretExtractorRBAC creates RBAC for secret extractor
func (c *Converter) createSecretExtractorRBAC() *RBAC {
	return &RBAC{
		Role: &Role{
			Name: "supabase-secret-reader",
			Rules: []PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get", "list"},
				},
			},
		},
		RoleBinding: &RoleBinding{
			Name: "supabase-secret-reader-binding",
			RoleRef: RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "supabase-secret-reader",
			},
			Subjects: []Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "supabase-secret-extractor",
					Namespace: "supabase",
				},
			},
		},
	}
}

// sanitizeName sanitizes a name for Kubernetes
func (c *Converter) sanitizeName(name string) string {
	// Replace invalid characters with dashes
	name = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(name, "-")
	// Remove leading/trailing dashes
	name = strings.Trim(name, "-")
	// Lowercase
	name = strings.ToLower(name)
	return name
}

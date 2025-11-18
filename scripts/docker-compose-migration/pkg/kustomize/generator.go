package kustomize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/converter"
)

// Generator generates Kustomize manifests
type Generator struct {
	resources *converter.K8sResources
	outputDir string
}

// NewGenerator creates a new Kustomize generator
func NewGenerator(resources *converter.K8sResources, outputDir string) *Generator {
	return &Generator{
		resources: resources,
		outputDir: outputDir,
	}
}

// Generate generates the Kustomize structure
func (g *Generator) Generate() error {
	// Create directory structure
	baseDir := filepath.Join(g.outputDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Generate base manifests
	if err := g.generateBase(baseDir); err != nil {
		return fmt.Errorf("failed to generate base: %w", err)
	}

	// Generate overlays
	overlays := []string{"development", "staging", "production"}
	for _, overlay := range overlays {
		overlayDir := filepath.Join(g.outputDir, "overlays", overlay)
		if err := os.MkdirAll(overlayDir, 0755); err != nil {
			return fmt.Errorf("failed to create overlay directory: %w", err)
		}
		if err := g.generateOverlay(overlayDir, overlay); err != nil {
			return fmt.Errorf("failed to generate overlay %s: %w", overlay, err)
		}
	}

	return nil
}

// generateBase generates base Kustomize manifests
func (g *Generator) generateBase(baseDir string) error {
	// Generate namespace
	if err := g.generateNamespace(baseDir); err != nil {
		return err
	}

	// Generate ConfigMaps
	for _, cm := range g.resources.ConfigMaps {
		if err := g.generateConfigMap(baseDir, cm); err != nil {
			return err
		}
	}

	// Generate Secrets
	for _, secret := range g.resources.Secrets {
		if err := g.generateSecret(baseDir, secret); err != nil {
			return err
		}
	}

	// Generate Services
	for _, svc := range g.resources.Services {
		if err := g.generateService(baseDir, svc); err != nil {
			return err
		}
	}

	// Generate Deployments
	for _, deploy := range g.resources.Deployments {
		if err := g.generateDeployment(baseDir, deploy); err != nil {
			return err
		}
	}

	// Generate StatefulSets
	for _, ss := range g.resources.StatefulSets {
		if err := g.generateStatefulSet(baseDir, ss); err != nil {
			return err
		}
	}

	// Generate ServiceAccount
	if g.resources.ServiceAccount != nil {
		if err := g.generateServiceAccount(baseDir, g.resources.ServiceAccount); err != nil {
			return err
		}
	}

	// Generate Secret Extractor Job
	if g.resources.SecretExtractorJob != nil {
		if err := g.generateJob(baseDir, g.resources.SecretExtractorJob); err != nil {
			return err
		}
	}

	// Generate RBAC
	if g.resources.SecretExtractorRBAC != nil {
		if err := g.generateRBAC(baseDir, g.resources.SecretExtractorRBAC); err != nil {
			return err
		}
	}

	// Generate kustomization.yaml
	if err := g.generateKustomization(baseDir); err != nil {
		return err
	}

	return nil
}

// generateNamespace generates namespace manifest
func (g *Generator) generateNamespace(baseDir string) error {
	namespace := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": g.resources.Namespace.Name,
		},
	}

	return g.writeYAML(filepath.Join(baseDir, "namespace.yaml"), namespace)
}

// generateConfigMap generates ConfigMap manifest
func (g *Generator) generateConfigMap(baseDir string, cm *converter.ConfigMap) error {
	configMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      cm.Name,
			"namespace": g.resources.Namespace.Name,
		},
		"data": cm.Data,
	}

	filename := fmt.Sprintf("configmap-%s.yaml", cm.Name)
	return g.writeYAML(filepath.Join(baseDir, filename), configMap)
}

// generateSecret generates Secret manifest
func (g *Generator) generateSecret(baseDir string, secret *converter.Secret) error {
	secretManifest := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      secret.Name,
			"namespace": g.resources.Namespace.Name,
		},
		"type": secret.Type,
	}

	if len(secret.Data) > 0 {
		secretManifest["data"] = secret.Data
	}

	if len(secret.StringData) > 0 {
		secretManifest["stringData"] = secret.StringData
	}

	filename := fmt.Sprintf("secret-%s.yaml", secret.Name)
	return g.writeYAML(filepath.Join(baseDir, filename), secretManifest)
}

// generateService generates Service manifest
func (g *Generator) generateService(baseDir string, svc *converter.ServiceResource) error {
	ports := []map[string]interface{}{}
	for _, port := range svc.Ports {
		ports = append(ports, map[string]interface{}{
			"name":       port.Name,
			"port":       port.Port,
			"targetPort": port.TargetPort,
			"protocol":   port.Protocol,
		})
	}

	service := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name":      svc.Name,
			"namespace": g.resources.Namespace.Name,
			"labels":    svc.Labels,
		},
		"spec": map[string]interface{}{
			"type":     svc.Type,
			"selector": svc.Selector,
			"ports":    ports,
		},
	}

	filename := fmt.Sprintf("service-%s.yaml", svc.Name)
	return g.writeYAML(filepath.Join(baseDir, filename), service)
}

// generateDeployment generates Deployment manifest
func (g *Generator) generateDeployment(baseDir string, deploy *converter.Deployment) error {
	deployment := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      deploy.Name,
			"namespace": g.resources.Namespace.Name,
			"labels":    deploy.Labels,
		},
		"spec": map[string]interface{}{
			"replicas": deploy.Replicas,
			"selector": map[string]interface{}{
				"matchLabels": deploy.Selector,
			},
			"template": g.generatePodTemplate(deploy.Template),
		},
	}

	filename := fmt.Sprintf("deployment-%s.yaml", deploy.Name)
	return g.writeYAML(filepath.Join(baseDir, filename), deployment)
}

// generateStatefulSet generates StatefulSet manifest
func (g *Generator) generateStatefulSet(baseDir string, ss *converter.StatefulSet) error {
	statefulSet := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name":      ss.Name,
			"namespace": g.resources.Namespace.Name,
			"labels":    ss.Labels,
		},
		"spec": map[string]interface{}{
			"serviceName": ss.ServiceName,
			"replicas":    ss.Replicas,
			"selector": map[string]interface{}{
				"matchLabels": ss.Selector,
			},
			"template": g.generatePodTemplate(ss.Template),
		},
	}

	// Add volume claim templates
	if len(ss.VolumeClaimTemplates) > 0 {
		vcTemplates := []map[string]interface{}{}
		for _, vct := range ss.VolumeClaimTemplates {
			vcTemplate := map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": vct.Name,
				},
				"spec": map[string]interface{}{
					"accessModes": vct.AccessModes,
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": vct.Size,
						},
					},
				},
			}
			if vct.StorageClass != "" {
				vcTemplate["spec"].(map[string]interface{})["storageClassName"] = vct.StorageClass
			}
			vcTemplates = append(vcTemplates, vcTemplate)
		}
		statefulSet["spec"].(map[string]interface{})["volumeClaimTemplates"] = vcTemplates
	}

	filename := fmt.Sprintf("statefulset-%s.yaml", ss.Name)
	return g.writeYAML(filepath.Join(baseDir, filename), statefulSet)
}

// generatePodTemplate generates a Pod template
func (g *Generator) generatePodTemplate(template converter.PodTemplate) map[string]interface{} {
	podTemplate := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": template.Labels,
		},
		"spec": map[string]interface{}{
			"containers": g.generateContainers(template.Containers),
		},
	}

	if len(template.Annotations) > 0 {
		podTemplate["metadata"].(map[string]interface{})["annotations"] = template.Annotations
	}

	if len(template.Volumes) > 0 {
		podTemplate["spec"].(map[string]interface{})["volumes"] = g.generateVolumes(template.Volumes)
	}

	if len(template.InitContainers) > 0 {
		podTemplate["spec"].(map[string]interface{})["initContainers"] = g.generateContainers(template.InitContainers)
	}

	return podTemplate
}

// generateContainers generates container specs
func (g *Generator) generateContainers(containers []converter.Container) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, container := range containers {
		c := map[string]interface{}{
			"name":  container.Name,
			"image": container.Image,
		}

		if len(container.Command) > 0 {
			c["command"] = container.Command
		}

		if len(container.Args) > 0 {
			c["args"] = container.Args
		}

		if len(container.Env) > 0 {
			c["env"] = g.generateEnvVars(container.Env)
		}

		if len(container.Ports) > 0 {
			c["ports"] = g.generateContainerPorts(container.Ports)
		}

		if len(container.VolumeMounts) > 0 {
			c["volumeMounts"] = g.generateVolumeMounts(container.VolumeMounts)
		}

		if container.LivenessProbe != nil {
			c["livenessProbe"] = g.generateProbe(container.LivenessProbe)
		}

		if container.ReadinessProbe != nil {
			c["readinessProbe"] = g.generateProbe(container.ReadinessProbe)
		}

		if container.Resources != nil {
			c["resources"] = container.Resources
		}

		if container.SecurityContext != nil {
			c["securityContext"] = g.generateSecurityContext(container.SecurityContext)
		}

		result = append(result, c)
	}
	return result
}

// generateEnvVars generates environment variables
func (g *Generator) generateEnvVars(envVars []converter.EnvVar) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, env := range envVars {
		e := map[string]interface{}{
			"name": env.Name,
		}

		if env.Value != "" {
			e["value"] = env.Value
		} else if env.ValueFrom != nil {
			valueFrom := map[string]interface{}{}
			if env.ValueFrom.SecretKeyRef != nil {
				valueFrom["secretKeyRef"] = map[string]interface{}{
					"name": env.ValueFrom.SecretKeyRef.Name,
					"key":  env.ValueFrom.SecretKeyRef.Key,
				}
			} else if env.ValueFrom.ConfigMapKeyRef != nil {
				valueFrom["configMapKeyRef"] = map[string]interface{}{
					"name": env.ValueFrom.ConfigMapKeyRef.Name,
					"key":  env.ValueFrom.ConfigMapKeyRef.Key,
				}
			} else if env.ValueFrom.FieldRef != nil {
				valueFrom["fieldRef"] = map[string]interface{}{
					"fieldPath": env.ValueFrom.FieldRef.FieldPath,
				}
			}
			e["valueFrom"] = valueFrom
		}

		result = append(result, e)
	}
	return result
}

// generateContainerPorts generates container ports
func (g *Generator) generateContainerPorts(ports []converter.ContainerPort) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, port := range ports {
		result = append(result, map[string]interface{}{
			"name":          port.Name,
			"containerPort": port.ContainerPort,
			"protocol":      port.Protocol,
		})
	}
	return result
}

// generateVolumeMounts generates volume mounts
func (g *Generator) generateVolumeMounts(mounts []converter.VolumeMount) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, mount := range mounts {
		m := map[string]interface{}{
			"name":      mount.Name,
			"mountPath": mount.MountPath,
		}
		if mount.SubPath != "" {
			m["subPath"] = mount.SubPath
		}
		if mount.ReadOnly {
			m["readOnly"] = true
		}
		result = append(result, m)
	}
	return result
}

// generateVolumes generates volumes
func (g *Generator) generateVolumes(volumes []converter.Volume) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, vol := range volumes {
		v := map[string]interface{}{
			"name": vol.Name,
		}

		if vol.VolumeSource.ConfigMap != nil {
			v["configMap"] = map[string]interface{}{
				"name": vol.VolumeSource.ConfigMap.Name,
			}
		} else if vol.VolumeSource.Secret != nil {
			v["secret"] = map[string]interface{}{
				"secretName": vol.VolumeSource.Secret.SecretName,
			}
		} else if vol.VolumeSource.PersistentVolumeClaim != nil {
			v["persistentVolumeClaim"] = map[string]interface{}{
				"claimName": vol.VolumeSource.PersistentVolumeClaim.ClaimName,
			}
		} else if vol.VolumeSource.EmptyDir != nil {
			emptyDir := map[string]interface{}{}
			if vol.VolumeSource.EmptyDir.Medium != "" {
				emptyDir["medium"] = vol.VolumeSource.EmptyDir.Medium
			}
			v["emptyDir"] = emptyDir
		} else if vol.VolumeSource.HostPath != nil {
			v["hostPath"] = map[string]interface{}{
				"path": vol.VolumeSource.HostPath.Path,
				"type": vol.VolumeSource.HostPath.Type,
			}
		}

		result = append(result, v)
	}
	return result
}

// generateProbe generates a probe
func (g *Generator) generateProbe(probe *converter.Probe) map[string]interface{} {
	p := map[string]interface{}{}

	if probe.Exec != nil {
		p["exec"] = map[string]interface{}{
			"command": probe.Exec.Command,
		}
	} else if probe.HTTPGet != nil {
		p["httpGet"] = map[string]interface{}{
			"path":   probe.HTTPGet.Path,
			"port":   probe.HTTPGet.Port,
			"scheme": probe.HTTPGet.Scheme,
		}
	} else if probe.TCPSocket != nil {
		p["tcpSocket"] = map[string]interface{}{
			"port": probe.TCPSocket.Port,
		}
	}

	if probe.InitialDelaySeconds > 0 {
		p["initialDelaySeconds"] = probe.InitialDelaySeconds
	}
	if probe.PeriodSeconds > 0 {
		p["periodSeconds"] = probe.PeriodSeconds
	}
	if probe.TimeoutSeconds > 0 {
		p["timeoutSeconds"] = probe.TimeoutSeconds
	}
	if probe.SuccessThreshold > 0 {
		p["successThreshold"] = probe.SuccessThreshold
	}
	if probe.FailureThreshold > 0 {
		p["failureThreshold"] = probe.FailureThreshold
	}

	return p
}

// generateSecurityContext generates a security context
func (g *Generator) generateSecurityContext(sc *converter.SecurityContext) map[string]interface{} {
	result := map[string]interface{}{}

	if sc.Privileged != nil {
		result["privileged"] = *sc.Privileged
	}
	if sc.RunAsUser != nil {
		result["runAsUser"] = *sc.RunAsUser
	}
	if sc.RunAsNonRoot != nil {
		result["runAsNonRoot"] = *sc.RunAsNonRoot
	}
	if sc.ReadOnlyRootFilesystem != nil {
		result["readOnlyRootFilesystem"] = *sc.ReadOnlyRootFilesystem
	}
	if sc.Capabilities != nil {
		caps := map[string]interface{}{}
		if len(sc.Capabilities.Add) > 0 {
			caps["add"] = sc.Capabilities.Add
		}
		if len(sc.Capabilities.Drop) > 0 {
			caps["drop"] = sc.Capabilities.Drop
		}
		result["capabilities"] = caps
	}

	return result
}

// generateServiceAccount generates ServiceAccount manifest
func (g *Generator) generateServiceAccount(baseDir string, sa *converter.ServiceAccount) error {
	serviceAccount := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]interface{}{
			"name":      sa.Name,
			"namespace": g.resources.Namespace.Name,
		},
	}

	return g.writeYAML(filepath.Join(baseDir, "serviceaccount.yaml"), serviceAccount)
}

// generateJob generates Job manifest
func (g *Generator) generateJob(baseDir string, job *converter.Job) error {
	jobManifest := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name":      job.Name,
			"namespace": g.resources.Namespace.Name,
			"labels":    job.Labels,
		},
		"spec": map[string]interface{}{
			"template": g.generatePodTemplate(job.Template),
		},
	}

	if len(job.Annotations) > 0 {
		jobManifest["metadata"].(map[string]interface{})["annotations"] = job.Annotations
	}

	return g.writeYAML(filepath.Join(baseDir, "secret-extractor-job.yaml"), jobManifest)
}

// generateRBAC generates RBAC manifests
func (g *Generator) generateRBAC(baseDir string, rbac *converter.RBAC) error {
	// Generate Role
	role := map[string]interface{}{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "Role",
		"metadata": map[string]interface{}{
			"name":      rbac.Role.Name,
			"namespace": g.resources.Namespace.Name,
		},
		"rules": []map[string]interface{}{},
	}

	for _, rule := range rbac.Role.Rules {
		role["rules"] = append(role["rules"].([]map[string]interface{}), map[string]interface{}{
			"apiGroups": rule.APIGroups,
			"resources": rule.Resources,
			"verbs":     rule.Verbs,
		})
	}

	// Generate RoleBinding
	roleBinding := map[string]interface{}{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "RoleBinding",
		"metadata": map[string]interface{}{
			"name":      rbac.RoleBinding.Name,
			"namespace": g.resources.Namespace.Name,
		},
		"roleRef": map[string]interface{}{
			"apiGroup": rbac.RoleBinding.RoleRef.APIGroup,
			"kind":     rbac.RoleBinding.RoleRef.Kind,
			"name":     rbac.RoleBinding.RoleRef.Name,
		},
		"subjects": []map[string]interface{}{},
	}

	for _, subject := range rbac.RoleBinding.Subjects {
		s := map[string]interface{}{
			"kind": subject.Kind,
			"name": subject.Name,
		}
		if subject.Namespace != "" {
			s["namespace"] = subject.Namespace
		}
		roleBinding["subjects"] = append(roleBinding["subjects"].([]map[string]interface{}), s)
	}

	// Write both to one file
	combined := []interface{}{role, roleBinding}
	return g.writeYAMLMulti(filepath.Join(baseDir, "secret-extractor-rbac.yaml"), combined)
}

// generateKustomization generates kustomization.yaml
func (g *Generator) generateKustomization(baseDir string) error {
	// Collect all resource files
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return err
	}

	resources := []string{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") && file.Name() != "kustomization.yaml" {
			resources = append(resources, file.Name())
		}
	}

	kustomization := map[string]interface{}{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"namespace":  g.resources.Namespace.Name,
		"resources":  resources,
	}

	return g.writeYAML(filepath.Join(baseDir, "kustomization.yaml"), kustomization)
}

// generateOverlay generates an overlay
func (g *Generator) generateOverlay(overlayDir string, overlay string) error {
	// Create kustomization.yaml for overlay
	kustomization := map[string]interface{}{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"resources": []string{
			"../../base",
		},
		"namePrefix": fmt.Sprintf("%s-", overlay),
	}

	// Add patches based on overlay type
	if overlay == "production" {
		kustomization["replicas"] = []map[string]interface{}{
			{
				"name":  "supabase-kong",
				"count": 2,
			},
			{
				"name":  "supabase-auth",
				"count": 2,
			},
		}
	}

	return g.writeYAML(filepath.Join(overlayDir, "kustomization.yaml"), kustomization)
}

// writeYAML writes a single YAML document to a file
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

// writeYAMLMulti writes multiple YAML documents to a file
func (g *Generator) writeYAMLMulti(filename string, data []interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	for i, doc := range data {
		if i > 0 {
			if _, err := file.WriteString("---\n"); err != nil {
				return err
			}
		}
		if err := encoder.Encode(doc); err != nil {
			return err
		}
	}

	return nil
}

package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GetPodEnvVar retrieves an environment variable from a pod
func (c *Client) GetPodEnvVar(ctx context.Context, podName, containerName, envVar string) (string, error) {
	output, err := c.ExecInPod(ctx, podName, containerName,
		[]string{"sh", "-c", fmt.Sprintf("echo $%s", envVar)})
	if err != nil {
		return "", fmt.Errorf("failed to get env var %s: %w", envVar, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetPodEnvVars retrieves multiple environment variables from a pod
func (c *Client) GetPodEnvVars(ctx context.Context, podName, containerName string, envVars []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, envVar := range envVars {
		value, err := c.GetPodEnvVar(ctx, podName, containerName, envVar)
		if err != nil {
			// Log error but continue with other vars
			continue
		}
		if value != "" {
			result[envVar] = value
		}
	}

	return result, nil
}

// GetAllPodEnvVars retrieves all environment variables from a pod
func (c *Client) GetAllPodEnvVars(ctx context.Context, podName, containerName string) (map[string]string, error) {
	output, err := c.ExecInPod(ctx, podName, containerName, []string{"env"})
	if err != nil {
		return nil, fmt.Errorf("failed to get env vars: %w", err)
	}

	result := make(map[string]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result, nil
}

// WaitForPodsReady waits for all pods matching a label selector to be ready
func (c *Client) WaitForPodsReady(ctx context.Context, labelSelector string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pods to be ready")
		case <-ticker.C:
			pods, err := c.ListPods(ctx, labelSelector)
			if err != nil {
				continue
			}

			if len(pods.Items) == 0 {
				continue
			}

			allReady := true
			for _, pod := range pods.Items {
				ready := false
				for _, condition := range pod.Status.Conditions {
					if condition.Type == "Ready" && condition.Status == "True" {
						ready = true
						break
					}
				}
				if !ready {
					allReady = false
					break
				}
			}

			if allReady {
				return nil
			}
		}
	}
}

// FindFirstReadyPod finds the first ready pod matching the label selector
func (c *Client) FindFirstReadyPod(ctx context.Context, labelSelector string) (string, string, error) {
	pod, err := c.FindPodByLabel(ctx, labelSelector)
	if err != nil {
		return "", "", err
	}

	// Get the first container name
	if len(pod.Spec.Containers) == 0 {
		return "", "", fmt.Errorf("pod %s has no containers", pod.Name)
	}

	return pod.Name, pod.Spec.Containers[0].Name, nil
}

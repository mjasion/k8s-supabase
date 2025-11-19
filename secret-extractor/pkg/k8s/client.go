package k8s

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// Client wraps a Kubernetes client with helper methods
type Client struct {
	clientset *kubernetes.Clientset
	namespace string
	config    *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient(namespace string) (*Client, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset: clientset,
		namespace: namespace,
		config:    config,
	}, nil
}

// GetPod retrieves a pod by name
func (c *Client) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListPods lists pods with the given label selector
func (c *Client) ListPods(ctx context.Context, labelSelector string) (*corev1.PodList, error) {
	return c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

// ExecInPod executes a command in a pod and returns the output
func (c *Client) ExecInPod(ctx context.Context, podName, containerName string, command []string) ([]byte, error) {
	req := c.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(c.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("exec failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ReadFileFromPod reads a file from a pod
func (c *Client) ReadFileFromPod(ctx context.Context, podName, containerName, filePath string) ([]byte, error) {
	return c.ExecInPod(ctx, podName, containerName, []string{"cat", filePath})
}

// WaitForPodReady waits for a pod to be ready
func (c *Client) WaitForPodReady(ctx context.Context, podName string) error {
	watch, err := c.clientset.CoreV1().Pods(c.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
	})
	if err != nil {
		return fmt.Errorf("failed to watch pod: %w", err)
	}
	defer watch.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watch.ResultChan():
			if event.Object == nil {
				return fmt.Errorf("watch channel closed unexpectedly")
			}
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			// Check if pod is ready
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return nil
				}
			}
		}
	}
}

// FindPodByLabel finds the first pod matching the label selector
func (c *Client) FindPodByLabel(ctx context.Context, labelSelector string) (*corev1.Pod, error) {
	pods, err := c.ListPods(ctx, labelSelector)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found with selector: %s", labelSelector)
	}

	// Return the first ready pod
	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return &pod, nil
			}
		}
	}

	// If no ready pod, return the first one
	return &pods.Items[0], nil
}

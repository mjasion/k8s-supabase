package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistExtractedSecrets creates or updates a Kubernetes Secret with extracted secrets
func (c *Client) PersistExtractedSecrets(
	ctx context.Context,
	name string,
	secrets map[string]string,
) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "supabase",
				"app.kubernetes.io/component":  "extracted-secrets",
				"app.kubernetes.io/managed-by": "secret-extractor",
			},
			Annotations: map[string]string{
				"supabase.io/extracted-at": time.Now().Format(time.RFC3339),
			},
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: secrets,
	}

	// Check if secret exists
	existing, err := c.clientset.CoreV1().Secrets(c.namespace).Get(
		ctx, name, metav1.GetOptions{})

	if err == nil {
		// Secret exists - update it
		existing.StringData = secret.StringData
		existing.Annotations = secret.Annotations
		_, err = c.clientset.CoreV1().Secrets(c.namespace).Update(
			ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Secret doesn't exist - create it
	_, err = c.clientset.CoreV1().Secrets(c.namespace).Create(
		ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetSecret retrieves a secret by name
func (c *Client) GetSecret(ctx context.Context, name string) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// SecretExists checks if a secret exists
func (c *Client) SecretExists(ctx context.Context, name string) (bool, error) {
	_, err := c.GetSecret(ctx, name)
	if err == nil {
		return true, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

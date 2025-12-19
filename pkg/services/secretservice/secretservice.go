package secretservice

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretService handles secret operations for Talos clusters
type SecretService struct {
	client.Client
}

// NewSecretService creates a new secret service
func NewSecretService(c client.Client) *SecretService {
	return &SecretService{
		Client: c,
	}
}

// GetTalosSecret retrieves the Talos secret for a cluster
func (s *SecretService) GetTalosSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := s.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, secret)
	return secret, err
}

// CreateTalosSecret creates a new Talos secret for a cluster
func (s *SecretService) CreateTalosSecret(ctx context.Context, name, namespace string, labels map[string]string, annotations map[string]string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	return s.Create(ctx, secret)
}

// UpdateTalosSecret updates an existing Talos secret for a cluster
func (s *SecretService) UpdateTalosSecret(ctx context.Context, secret *corev1.Secret) error {
	return s.Update(ctx, secret)
}

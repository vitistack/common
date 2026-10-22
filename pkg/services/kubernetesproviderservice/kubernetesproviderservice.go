package kubernetesproviderservice

import (
	"context"

	"fmt"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/unstructuredutil"
	vitistackv1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// kubernetesProviderGVR is the GroupVersionResource for vitistack.io/v1alpha1 KubernetesProvider
var kubernetesProviderGVR = schema.GroupVersionResource{
	Group:    "vitistack.io",
	Version:  "v1alpha1",
	Resource: "kubernetesproviders",
}

// getProviderNames extracts names from a list of KubernetesProviders
func GetProviderNames(providers []*vitistackv1alpha1.KubernetesProvider) []string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name
	}
	return names
}

// ListAllKubernetesProviders lists all KubernetesProviders in the cluster
func ListAllKubernetesProviders(ctx context.Context) ([]*vitistackv1alpha1.KubernetesProvider, error) {
	unstructuredList, err := k8sclient.DynamicClient.Resource(kubernetesProviderGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list KubernetesProviders: %w", err)
	}

	providers := make([]*vitistackv1alpha1.KubernetesProvider, 0, len(unstructuredList.Items))
	for i := range unstructuredList.Items {
		provider, err := unstructuredutil.KubernetesProviderFromUnstructured(&unstructuredList.Items[i])
		if err != nil {
			vlog.Warn("Failed to convert KubernetesProvider, skipping",
				"name", unstructuredList.Items[i].GetName(),
				"error", err)
			continue
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

// ListKubernetesProvidersByType lists all KubernetesProviders with the specified providerType
func ListKubernetesProvidersByType(ctx context.Context, providerType string) ([]*vitistackv1alpha1.KubernetesProvider, error) {
	allProviders, err := ListAllKubernetesProviders(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]*vitistackv1alpha1.KubernetesProvider, 0)
	for _, provider := range allProviders {
		if provider.Spec.ProviderType == providerType {
			filtered = append(filtered, provider)
		}
	}

	return filtered, nil
}

// KubernetesProviderExistsByName checks if a KubernetesProvider exists with the given name
func KubernetesProviderExistsByName(ctx context.Context, name string) (bool, error) {
	_, err := k8sclient.DynamicClient.Resource(kubernetesProviderGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetKubernetesProviderByName fetches a KubernetesProvider by name
func GetKubernetesProviderByName(ctx context.Context, name string) (*vitistackv1alpha1.KubernetesProvider, error) {
	unstructured, err := k8sclient.DynamicClient.Resource(kubernetesProviderGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get KubernetesProvider %q: %w", name, err)
	}

	provider, err := unstructuredutil.KubernetesProviderFromUnstructured(unstructured)
	if err != nil {
		return nil, fmt.Errorf("failed to convert KubernetesProvider: %w", err)
	}

	return provider, nil
}

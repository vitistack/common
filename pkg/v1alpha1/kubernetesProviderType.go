package v1alpha1

type KubernetesProviderType string

const (
	KubernetesProviderTypeTalos KubernetesProviderType = "talos"
	KubernetesProviderTypeAKS   KubernetesProviderType = "aks"
)

func (pt KubernetesProviderType) IsValid() bool {
	switch pt {
	case KubernetesProviderTypeTalos:
		return true
	default:
		return false
	}
}

func ValidKubernetesProviderTypes() []KubernetesProviderType {
	return []KubernetesProviderType{
		KubernetesProviderTypeTalos,
	}
}

func DefaultKubernetesProviderType() KubernetesProviderType {
	return KubernetesProviderTypeTalos
}

func KubernetesProviderTypeValues() []string {
	values := make([]string, len(ValidKubernetesProviderTypes()))
	for i, pt := range ValidKubernetesProviderTypes() {
		values[i] = string(pt)
	}
	return values
}

func (pt KubernetesProviderType) String() string {
	return string(pt)
}

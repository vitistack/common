package unstructuredutil

import (
	"fmt"

	v1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// defaultScheme contains the project API types registered for GVK resolution.
var defaultScheme = func() *runtime.Scheme {
	s := runtime.NewScheme()
	// Register our API group.
	_ = v1alpha1.AddToScheme(s)
	return s
}()

// ToUnstructured converts a typed Kubernetes object into *unstructured.Unstructured.
// If APIVersion/Kind are missing on the object, they are inferred from the scheme.
func ToUnstructured(obj runtime.Object) (*metav1unstructured.Unstructured, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	// Ensure GVK is set when possible so apiversion/kind appear in the output map.
	if gvks, _, err := defaultScheme.ObjectKinds(obj); err == nil && len(gvks) > 0 {
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	}

	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &metav1unstructured.Unstructured{Object: m}, nil
}

// FromUnstructured converts an *unstructured.Unstructured into the provided typed object.
// 'into' must be a pointer to a struct for the target kind (e.g., *v1alpha1.Machine).
func FromUnstructured(u *metav1unstructured.Unstructured, into runtime.Object) error {
	if u == nil {
		return fmt.Errorf("nil unstructured")
	}
	if into == nil {
		return fmt.Errorf("nil target object")
	}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, into)
}

// Typed helpers for each CRD kind in this module.

func MachineToUnstructured(in *v1alpha1.Machine) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func MachineFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.Machine, error) {
	out := new(v1alpha1.Machine)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func MachineProviderToUnstructured(in *v1alpha1.MachineProvider) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func MachineProviderFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.MachineProvider, error) {
	out := new(v1alpha1.MachineProvider)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func KubernetesProviderToUnstructured(in *v1alpha1.KubernetesProvider) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func KubernetesProviderFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.KubernetesProvider, error) {
	out := new(v1alpha1.KubernetesProvider)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func KubernetesClusterToUnstructured(in *v1alpha1.KubernetesCluster) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func KubernetesClusterFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.KubernetesCluster, error) {
	out := new(v1alpha1.KubernetesCluster)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func NetworkConfigurationToUnstructured(in *v1alpha1.NetworkConfiguration) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func NetworkConfigurationFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.NetworkConfiguration, error) {
	out := new(v1alpha1.NetworkConfiguration)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func NetworkNamespaceToUnstructured(in *v1alpha1.NetworkNamespace) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func NetworkNamespaceFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.NetworkNamespace, error) {
	out := new(v1alpha1.NetworkNamespace)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func VitistackToUnstructured(in *v1alpha1.Vitistack) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func VitistackFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.Vitistack, error) {
	out := new(v1alpha1.Vitistack)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

func LoadBalancerToUnstructured(in *v1alpha1.ControlPlaneVirtualSharedIP) (*metav1unstructured.Unstructured, error) {
	return ToUnstructured(in)
}

func LoadBalancerFromUnstructured(u *metav1unstructured.Unstructured) (*v1alpha1.ControlPlaneVirtualSharedIP, error) {
	out := new(v1alpha1.ControlPlaneVirtualSharedIP)
	if err := FromUnstructured(u, out); err != nil {
		return nil, err
	}
	return out, nil
}

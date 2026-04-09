package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageConfigClass is the Schema for the StorageConfigClass API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=storageconfigclasses,scope=Cluster,shortName=scc
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Enabled",type=string,JSONPath=`.spec.enabled`

type StorageConfigClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StorageConfigClassSpec `json:"spec,omitempty"`

	Status StorageConfigClassStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// StorageConfigClassList contains a list of StorageConfigClass
type StorageConfigClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageConfigClass `json:"items"`
}

type StorageConfigClassSpec struct {
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Required
	Version string `json:"version,omitempty"`

	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`

	// +kubebuilder:validation:Optional
	Data map[string]string `json:"data,omitempty"`
}

type StorageConfigClassStatus struct {
	Name string `json:"name,omitempty"`
}

func Init() {
	SchemeBuilder.Register(&StorageConfigClass{}, &StorageConfigClassList{})
}

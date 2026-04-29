package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterStorageClass is the Schema for the ClusterStorageClass API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterstorageclasses,scope=Cluster,shortName=csc
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Enabled",type=string,JSONPath=`.spec.enabled`

type ClusterStorageClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterStorageClassSpec `json:"spec,omitempty"`

	Status ClusterStorageClassStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClusterStorageClassList contains a list of ClusterStorageClass
type ClusterStorageClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterStorageClass `json:"items"`
}

type ClusterStorageClassSpec struct {
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Required
	Version string `json:"version,omitempty"`

	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`

	// +kubebuilder:validation:Optional
	Data map[string]string `json:"data,omitempty"`
}

type ClusterStorageClassStatus struct {
	Name string `json:"name,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ClusterStorageClass{}, &ClusterStorageClassList{})
}

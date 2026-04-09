package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterStorage is the Schema for the ClusterStorage API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterstorages,scope=Namespaced,shortName=cls
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.clusterId`
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.spec.clusterNamespace`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`

type ClusterStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterStorageSpec `json:"spec,omitempty"`

	Status ClusterStorageStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClusterStorageList contains a list of ClusterStorage
type ClusterStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterStorage `json:"items"`
}

type ClusterStorageSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	ClusterId string `json:"clusterId,omitempty"`

	// +kubebuilder:validation:Required
	ClusterNamespace string `json:"clusterNamespace,omitempty"`

	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`

	// +kubebuilder:validation:Required
	StorageConfigClass string `json:"storageConfigClass,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	ReuseExisting bool `json:"reuseExisting,omitempty"`

	// +kubebuilder:validation:Optional
	ExistingRef string `json:"existingRef,omitempty"`
}

type ClusterStorageStatus struct {
	Name string `json:"name,omitempty"`

	Phase string `json:"phase,omitempty"`

	Message string `json:"message,omitempty"`

	Secret secretStatus `json:"secret,omitempty"`

	GuestResource GuestResourceStatus `json:"guestResource,omitempty"`
}

type GuestResourceStatus struct {
	Condition string `json:"condition,omitempty"`
}

type secretStatus struct {
	Name string `json:"name,omitempty"`

	Condition string `json:"condition,omitempty"`

	Message string `json:"message,omitempty"`
}

const (
	secretConditionCreated = "Created"
	secretConditionError   = "Error"
	secretConditionPending = "Pending"
	secretConditionReady   = "Ready"
)

const (
	ClusterStoragePhasePending      = "Pending"
	ClusterStoragePhaseInitializing = "Initializing"
	ClusterStoragePhaseDeploying    = "Deploying"
	ClusterStoragePhaseFailed       = "Failed"
	ClusterStoragePhaseReady        = "Ready"
)

func init() {
	SchemeBuilder.Register(&ClusterStorage{}, &ClusterStorageList{})
}

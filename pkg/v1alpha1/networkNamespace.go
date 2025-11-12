package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkNamespace is the Schema for the NetworkNamespace API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=networknamespaces,scope=Namespaced,shortName=nn
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.clusterIdentifier`
// +kubebuilder:printcolumn:name="DatacenterIdentifier",type=string,JSONPath=`.spec.datacenterIdentifier`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Created",type=string,JSONPath=`.status.created`,description="Creation Timestamp"
type NetworkNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkNamespaceSpec   `json:"spec,omitempty"`
	Status NetworkNamespaceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkNamespaceList contains a list of NetworkNamespace
type NetworkNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkNamespace `json:"items"`
}

type NetworkNamespaceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	DatacenterIdentifier string `json:"datacenterIdentifier,omitempty"` // <country>-<region>-<availability zone> ex: no-west-az1

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"` // <unique name per datacenter> ex: my-namespace

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	ClusterIdentifier string `json:"clusterIdentifier,omitempty"` // <unique name per availability zone> ex: my-name

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	Provider string `json:"provider,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9._-]+$`
	Environment string `json:"environment,omitempty"`
}

type NetworkNamespaceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Phase      string             `json:"phase,omitempty"`
	Status     string             `json:"status,omitempty"`
	Message    string             `json:"message,omitempty"`
	Created    metav1.Time        `json:"created,omitempty"`

	DataCenterIdentifier string `json:"datacenterIdentifier,omitempty"`
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"`
	NamespaceID          string `json:"namespaceId,omitempty"`
	ClusterIdentifier    string `json:"clusterIdentifier,omitempty"`
	IPv4Prefix           string `json:"ipv4Prefix,omitempty"`
	IPv6Prefix           string `json:"ipv6Prefix,omitempty"`
	IPv4EgressIP         string `json:"ipv4EgressIp,omitempty"`
	IPv6EgressIP         string `json:"ipv6EgressIp,omitempty"`
	VlanID               int    `json:"vlanId,omitempty"`

	AssociatedKubernetesClusterIDs []string `json:"associatedKubernetesClusterIds,omitempty"`
}

func init() {
	SchemeBuilder.Register(&NetworkNamespace{}, &NetworkNamespaceList{})
}

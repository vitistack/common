package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneVirtualSharedIP is the Schema for the ControlPlaneVirtualSharedIP API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=controlplanevirtualsharedips,scope=Namespaced,shortName=lb
// +kubebuilder:printcolumn:name="DatacenterIdentifier",type=string,JSONPath=`.spec.datacenterIdentifier`
// +kubebuilder:printcolumn:name="ClusterIdentifier",type=string,JSONPath=`.spec.clusterIdentifier`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Created",type=string,JSONPath=`.status.created`,description="Creation Timestamp"
type ControlPlaneVirtualSharedIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ControlPlaneVirtualSharedIPSpec `json:"spec,omitempty"`

	Status ControlPlaneVirtualSharedIPStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ControlPlaneVirtualSharedIPList contains a list of ControlPlaneVirtualSharedIP
type ControlPlaneVirtualSharedIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlaneVirtualSharedIP `json:"items"`
}

type ControlPlaneVirtualSharedIPSpec struct {
	// +kubebuilder:validation:Required
	DatacenterIdentifier string `json:"datacenterIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	NetworkNamespaceIdentifier string `json:"networkNamespaceIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	ClusterIdentifier string `json:"clusterIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	Provider string `json:"provider,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=first-alive
	// +kubebuilder:validation:Enum=round-robin;least-session;first-alive
	// round-robin, least-session, first-alive
	Method string `json:"method,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9._-]+$`
	Environment string `json:"environment,omitempty"`

	PoolMembers []string `json:"poolMembers,omitempty"`
}

type ControlPlaneVirtualSharedIPStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Phase      string             `json:"phase,omitempty"`
	Status     string             `json:"status,omitempty"`
	Message    string             `json:"message,omitempty"`
	Created    metav1.Time        `json:"created,omitempty"`

	DatacenterIdentifier       string   `json:"datacenterIdentifier,omitempty"`
	SupervisorIdentifier       string   `json:"supervisorIdentifier,omitempty"`
	ClusterIdentifier          string   `json:"clusterIdentifier,omitempty"`
	LoadBalancerIps            []string `json:"loadBalancerIps,omitempty"`
	Method                     string   `json:"method,omitempty"`
	PoolMembers                []string `json:"poolMembers,omitempty"`
	NetworkNamespaceIdentifier string   `json:"networkNamespaceIdentifier,omitempty"`
	Environment                string   `json:"environment,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ControlPlaneVirtualSharedIP{}, &ControlPlaneVirtualSharedIPList{})
}

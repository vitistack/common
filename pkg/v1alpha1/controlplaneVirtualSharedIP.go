package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneVirtualSharedIP is the Schema for the ControlPlaneVirtualSharedIP API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:ac:generate=true
// +kubebuilder:resource:path=controlplanevirtualsharedips,scope=Namespaced,shortName=lb;cpvip
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Datacenter ID",type=string,JSONPath=`.status.datacenterIdentifier`
// +kubebuilder:printcolumn:name="Control Plane VIP",type=string,JSONPath=`.status.loadBalancerIps[*]`
// +kubebuilder:printcolumn:name="Cluster ID",type=string,JSONPath=`.spec.clusterIdentifier`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="ClusterProvisioned")].status`
// +kubebuilder:printcolumn:name="Created",type=date,JSONPath=`.metadata.creationTimestamp`,description="Creation Timestamp"
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="ClusterProvisioned")].reason`,priority=1
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.conditions[?(@.type=="ClusterProvisioned")].message`,priority=1
// +kubebuilder:printcolumn:name="Transition time",type=date,JSONPath=`.status.conditions[?(@.type=="ClusterProvisioned")].lastTransitionTime`,priority=1
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
	// +kubebuilder:validation:Optional
	DatacenterIdentifier string `json:"datacenterIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	NetworkNamespaceIdentifier string `json:"networkNamespaceIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	ClusterIdentifier string `json:"clusterIdentifier,omitempty"`

	// +kubebuilder:validation:Optional
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

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	RetryCount         int   `json:"retryCount,omitempty"`

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

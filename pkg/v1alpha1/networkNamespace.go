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

	// IPAllocation defines how IP addresses are allocated within this NetworkNamespace.
	// When not set, the default behavior is DHCP-based allocation (backward compatible
	// with existing nms-operator + kea-operator flow).
	// +kubebuilder:validation:Optional
	IPAllocation *NetworkNamespaceIPAllocation `json:"ipAllocation,omitempty"`
}

// NetworkNamespaceIPAllocation configures the IP allocation method for a NetworkNamespace.
type NetworkNamespaceIPAllocation struct {
	// Type specifies the IP allocation method to use.
	// "dhcp" uses an external DHCP server for address assignment.
	// "static" uses a static IP operator to allocate from a defined range.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=dhcp;static
	Type IPAllocationType `json:"type"`

	// Provider identifies the operator or system that implements the allocation.
	// Examples: "kea" (Kea DHCP server), "static-ip-operator" or others
	// When empty, the default provider for the type is assumed.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	Provider string `json:"provider,omitempty"`

	// Static contains the IP pool configuration for static allocation.
	// Required when type is "static".
	// +kubebuilder:validation:Optional
	Static *StaticIPAllocationConfig `json:"static,omitempty"`

	// DHCP contains optional configuration overrides for DHCP-based allocation.
	// +kubebuilder:validation:Optional
	DHCP *DHCPAllocationConfig `json:"dhcp,omitempty"`
}

type NetworkNamespaceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Phase      string             `json:"phase,omitempty"`
	Status     string             `json:"status,omitempty"`
	Message    string             `json:"message,omitempty"`
	Created    metav1.Time        `json:"created,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	RetryCount         int   `json:"retryCount,omitempty"`

	DataCenterIdentifier string `json:"datacenterIdentifier,omitempty"`
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"`
	NamespaceID          string `json:"namespaceId,omitempty"`
	IPv4Prefix           string `json:"ipv4Prefix,omitempty"`
	IPv6Prefix           string `json:"ipv6Prefix,omitempty"`
	IPv4EgressIP         string `json:"ipv4EgressIp,omitempty"`
	IPv6EgressIP         string `json:"ipv6EgressIp,omitempty"`
	VlanID               int    `json:"vlanId,omitempty"`

	AssociatedKubernetesClusterIDs []string `json:"associatedKubernetesClusterIds,omitempty"`

	// IPAllocationStatus reports the current state of IP allocation
	// within this NetworkNamespace.
	// +kubebuilder:validation:Optional
	IPAllocationStatus *NetworkNamespaceIPAllocationStatus `json:"ipAllocationStatus,omitempty"`
}

// NetworkNamespaceIPAllocationStatus reports the observed IP allocation state.
type NetworkNamespaceIPAllocationStatus struct {
	// Type is the active IP allocation type.
	// +kubebuilder:validation:Optional
	Type IPAllocationType `json:"type,omitempty"`

	// Provider is the operator or system that performed the allocation.
	// +kubebuilder:validation:Optional
	Provider string `json:"provider,omitempty"`

	// AllocatedCount is the number of IP addresses currently allocated.
	// +kubebuilder:validation:Optional
	AllocatedCount int32 `json:"allocatedCount,omitempty"`

	// AvailableCount is the number of IP addresses available for allocation.
	// +kubebuilder:validation:Optional
	AvailableCount int32 `json:"availableCount,omitempty"`

	// TotalCount is the total number of IP addresses in the pool.
	// +kubebuilder:validation:Optional
	TotalCount int32 `json:"totalCount,omitempty"`

	// AllocatedIPs lists each currently allocated IP along with the
	// NetworkConfiguration that owns it.
	// +kubebuilder:validation:Optional
	AllocatedIPs []AllocatedIPEntry `json:"allocatedIPs,omitempty"`
}

// AllocatedIPEntry records a single IP allocation and the resource that owns it.
type AllocatedIPEntry struct {
	// IP is the allocated IPv4 address.
	IP string `json:"ip"`
	// NetworkConfiguration is the name of the NetworkConfiguration that owns this allocation.
	NetworkConfiguration string `json:"networkConfiguration"`
}

func init() {
	SchemeBuilder.Register(&NetworkNamespace{}, &NetworkNamespaceList{})
}

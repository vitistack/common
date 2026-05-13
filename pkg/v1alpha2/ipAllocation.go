package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Labels used on IPAllocation resources for efficient listing.
const (
	// LabelNetworkNamespace is the name of the NetworkNamespace this allocation belongs to.
	LabelNetworkNamespace = "vitistack.io/network-namespace"

	// LabelNetworkConfiguration is the name of the NetworkConfiguration that owns this allocation.
	LabelNetworkConfiguration = "vitistack.io/network-configuration"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAllocation represents a single IP address allocation within a NetworkNamespace.
// Each IPAllocation is owned (via ownerReference) by its NetworkConfiguration,
// so it is garbage-collected when the NC is deleted.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ipallocations,scope=Namespaced,shortName=ipa
// +kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.status.address`
// +kubebuilder:printcolumn:name="NetworkNamespace",type=string,JSONPath=`.spec.networkNamespaceName`
// +kubebuilder:printcolumn:name="NetworkConfiguration",type=string,JSONPath=`.spec.networkConfigurationName`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type IPAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPAllocationSpec   `json:"spec,omitempty"`
	Status IPAllocationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// IPAllocationList contains a list of IPAllocation
type IPAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAllocation `json:"items"`
}

// IPAllocationSpec defines the desired state of an IP allocation.
type IPAllocationSpec struct {
	// NetworkNamespaceName is the name of the NetworkNamespace this allocation
	// belongs to. Used to look up the IP pool configuration.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	NetworkNamespaceName string `json:"networkNamespaceName"`

	// NetworkConfigurationName is the name of the NetworkConfiguration that
	// requested this allocation.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	NetworkConfigurationName string `json:"networkConfigurationName"`

	// InterfaceName is the network interface name within the NetworkConfiguration
	// that this allocation is for. Supports multi-NIC VMs.
	// +kubebuilder:validation:Optional
	InterfaceName string `json:"interfaceName,omitempty"`

	// RequestedAddress allows requesting a specific IP address from the pool.
	// The allocator will try to honor this request if the address is available.
	// +kubebuilder:validation:Optional
	RequestedAddress string `json:"requestedAddress,omitempty"`
}

// IPAllocationStatus defines the observed state of an IP allocation.
type IPAllocationStatus struct {
	// Phase is the current lifecycle phase of this allocation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Allocated;Released;Error
	Phase IPAllocationPhase `json:"phase,omitempty"`

	// Address is the allocated IPv4 address (e.g. "10.0.2.50").
	// +kubebuilder:validation:Optional
	Address string `json:"address,omitempty"`

	// Gateway is the gateway address for the allocated subnet.
	// +kubebuilder:validation:Optional
	Gateway string `json:"gateway,omitempty"`

	// Prefix is the CIDR prefix length (e.g. 24 for a /24).
	// +kubebuilder:validation:Optional
	Prefix int `json:"prefix,omitempty"`

	// VlanID is the VLAN ID of the network segment.
	// +kubebuilder:validation:Optional
	VlanID int `json:"vlanId,omitempty"`

	// DNS is the list of DNS server addresses for this allocation.
	// +kubebuilder:validation:Optional
	DNS []string `json:"dns,omitempty"`

	// ExpiresAt is when this allocation expires (for static allocations with TTL).
	// +kubebuilder:validation:Optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// Message provides human-readable details about the current phase.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}

func init() {
	SchemeBuilder.Register(&IPAllocation{}, &IPAllocationList{})
}

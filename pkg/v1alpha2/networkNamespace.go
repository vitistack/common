package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkNamespace is the Schema for the NetworkNamespace API (v1alpha2).
//
// v1alpha2 separates network provisioning (where does the IP prefix come from?)
// from IP allocation (how are individual IPs assigned?). This enables pluggable
// IPAM back-ends and explicit provisioning readiness gating.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=networknamespaces,scope=Namespaced,shortName=nn
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Provisioning",type=string,JSONPath=`.spec.networkProvisioning.provider`
// +kubebuilder:printcolumn:name="DatacenterIdentifier",type=string,JSONPath=`.spec.datacenterIdentifier`
// +kubebuilder:printcolumn:name="ProvisioningPhase",type=string,JSONPath=`.status.provisioningPhase`
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

// NetworkNamespaceSpec defines the desired state of a NetworkNamespace.
type NetworkNamespaceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	DatacenterIdentifier string `json:"datacenterIdentifier,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"`

	// NetworkProvisioning defines how the network segment (IP prefix + VLAN) is
	// acquired. When not set, defaults to provider "nam" for backward compatibility.
	// +kubebuilder:validation:Optional
	NetworkProvisioning *NetworkProvisioning `json:"networkProvisioning,omitempty"`

	// IPAllocation defines how individual IP addresses are assigned within the
	// provisioned network. When not set, defaults to DHCP-based allocation.
	// +kubebuilder:validation:Optional
	IPAllocation *NetworkNamespaceIPAllocation `json:"ipAllocation,omitempty"`
}

// NetworkProvisioning configures the source of the network segment.
type NetworkProvisioning struct {
	// Provider identifies the system that provisions the network.
	// "nam" uses the Network Administration Management backend (default).
	// "manual" uses user-supplied configuration from the manual block.
	// Other values are reserved for future IPAM integrations.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=nam;manual
	// +kubebuilder:default=nam
	Provider NetworkProvisioningType `json:"provider"`

	// NAM contains configuration specific to NAM-based provisioning.
	// Reserved for future NAM-specific overrides.
	// +kubebuilder:validation:Optional
	NAM *NAMProvisioningConfig `json:"nam,omitempty"`

	// Manual contains user-supplied network configuration.
	// Required when provider is "manual".
	// +kubebuilder:validation:Optional
	Manual *ManualProvisioningConfig `json:"manual,omitempty"`
}

// NAMProvisioningConfig holds configuration specific to NAM-based network
// provisioning. Currently empty; reserved for future overrides.
type NAMProvisioningConfig struct{}

// ManualProvisioningConfig holds user-supplied network segment configuration.
type ManualProvisioningConfig struct {
	// IPv4CIDR is the subnet in CIDR notation (e.g. "10.0.2.0/24").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`
	IPv4CIDR string `json:"ipv4CIDR"`

	// IPv4Gateway is the default gateway address for the subnet.
	// +kubebuilder:validation:Optional
	IPv4Gateway string `json:"ipv4Gateway,omitempty"`

	// IPv6CIDR is the IPv6 subnet in CIDR notation.
	// +kubebuilder:validation:Optional
	IPv6CIDR string `json:"ipv6CIDR,omitempty"`

	// VlanID is the VLAN ID for the network segment.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	VlanID int `json:"vlanId,omitempty"`
}

// NetworkNamespaceIPAllocation configures the IP allocation method.
type NetworkNamespaceIPAllocation struct {
	// Type specifies the IP allocation method.
	// +kubebuilder:validation:Required
	Type IPAllocationType `json:"type"`

	// Provider identifies the operator that implements the allocation.
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

// NetworkNamespaceStatus defines the observed state of a NetworkNamespace.
type NetworkNamespaceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Phase      string             `json:"phase,omitempty"`
	Status     string             `json:"status,omitempty"`
	Message    string             `json:"message,omitempty"`
	Created    metav1.Time        `json:"created,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	RetryCount         int   `json:"retryCount,omitempty"`

	// ProvisioningPhase indicates whether the network segment has been
	// successfully provisioned. Downstream operators (kea, static-ip)
	// must wait for "Ready" before acting.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Ready;Error
	ProvisioningPhase ProvisioningPhase `json:"provisioningPhase,omitempty"`

	// Network fields populated by the provisioning layer.
	DataCenterIdentifier string `json:"datacenterIdentifier,omitempty"`
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"`
	NamespaceID          string `json:"namespaceId,omitempty"`
	IPv4Prefix           string `json:"ipv4Prefix,omitempty"`
	IPv6Prefix           string `json:"ipv6Prefix,omitempty"`
	IPv4EgressIP         string `json:"ipv4EgressIp,omitempty"`
	IPv6EgressIP         string `json:"ipv6EgressIp,omitempty"`
	VlanID               int    `json:"vlanId,omitempty"`

	AssociatedKubernetesClusterIDs []string `json:"associatedKubernetesClusterIds,omitempty"`

	// IPAllocationSummary reports aggregate IP allocation state. This is a
	// projection computed from IPAllocation resources, not the source of truth.
	// +kubebuilder:validation:Optional
	IPAllocationSummary *IPAllocationSummary `json:"ipAllocationSummary,omitempty"`
}

// IPAllocationSummary reports aggregate counts for IP allocation within
// a NetworkNamespace. Individual allocations are tracked in IPAllocation CRs.
type IPAllocationSummary struct {
	// Type is the active IP allocation type.
	Type IPAllocationType `json:"type,omitempty"`

	// Provider is the operator that performed the allocation.
	Provider string `json:"provider,omitempty"`

	// AllocatedCount is the number of IP addresses currently allocated.
	AllocatedCount int32 `json:"allocatedCount,omitempty"`

	// AvailableCount is the number of IP addresses available for allocation.
	AvailableCount int32 `json:"availableCount,omitempty"`

	// TotalCount is the total number of IP addresses in the pool.
	TotalCount int32 `json:"totalCount,omitempty"`
}

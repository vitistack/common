package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkNamespace is the Schema for the NetworkNamespace API (v1alpha3).
//
// v1alpha3 separates network provisioning (where does the IP prefix come from?)
// from IP allocation (how are individual IPs assigned?). This enables pluggable
// IPAM back-ends and explicit provisioning readiness gating.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:ac:generate=true
// +kubebuilder:resource:path=networknamespaces,scope=Namespaced,shortName=nn
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.networkProvisioning.provider`
// +kubebuilder:printcolumn:name="Datacenter ID",type=string,JSONPath=`.status.datacenterId`
// +kubebuilder:printcolumn:name="VLAN ID",type=string,JSONPath=`.status.vlanId`
// +kubebuilder:printcolumn:name="IPv4 Prefix",type=string,JSONPath=`.status.ipv4Prefix`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="NetworkProvisioned")].status`
// +kubebuilder:printcolumn:name="Created",type=date,JSONPath=`.metadata.creationTimestamp`,description="Creation Timestamp"
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

// IPAllocationType specifies the method used for IP address allocation within a NetworkNamespace.
// +kubebuilder:validation:Enum=dhcp;static
type IPAllocationType string

const (
	IPAllocationTypeDHCP   IPAllocationType = "dhcp"
	IPAllocationTypeStatic IPAllocationType = "static"
)

// NetworkProvisioningType identifies the system that provisions network
// segments (IP prefixes + VLANs) for a NetworkNamespace. Open string: the
// constants below are well-known values, not an exhaustive list.
type NetworkProvisioningType string

// ProvisioningPhase represents the lifecycle state of network provisioning.
// +kubebuilder:validation:Enum=Pending;Ready;Error
type ProvisioningPhase string

const (
	ProvisioningPhasePending ProvisioningPhase = "Pending"
	ProvisioningPhaseReady   ProvisioningPhase = "Ready"
	ProvisioningPhaseError   ProvisioningPhase = "Error"
)

// NetworkNamespaceSpec defines the desired state of a NetworkNamespace.
type NetworkNamespaceSpec struct {
	// NetworkProvisioning defines how the network segment (IP prefix + VLAN) is
	// acquired. When not set, defaults to provider "nam" for backward compatibility.
	// +kubebuilder:validation:Optional
	NetworkProvisioning *NetworkProvisioning `json:"networkProvisioning,omitempty"`

	// IPAllocation defines how individual IP addresses are assigned within the
	// provisioned network. When not set, defaults to DHCP-based allocation.
	// +kubebuilder:validation:Optional
	IPAllocation *NetworkNamespaceIPAllocation `json:"ipAllocation,omitempty"`

	// Egress requests egress addresses for the namespace.
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=network
	// +kubebuilder:validation:MaxItems=8
	Egress []EgressRequest `json:"egress,omitempty"`
}

// EgressRequest requests egress addresses on one network.
// +kubebuilder:validation:XValidation:rule="has(self.ipv4) || has(self.ipv6) || has(self.dynamicIPv4) || has(self.dynamicIPv6)",message="at least one of ipv4, ipv6, dynamicIPv4 or dynamicIPv6 must be set"
// +kubebuilder:validation:XValidation:rule="!has(self.ipv4) || self.ipv4.all(a, isIP(a) && ip(a).family() == 4)",message="ipv4 must contain IPv4 addresses"
// +kubebuilder:validation:XValidation:rule="!has(self.ipv6) || self.ipv6.all(a, isIP(a) && ip(a).family() == 6)",message="ipv6 must contain IPv6 addresses"
type EgressRequest struct {
	// Network is the egress network, e.g. "hnet" or "inet".
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=32
	Network string `json:"network"`

	// DynamicIPv4 is how many IPv4 addresses the provider should pick.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=16
	DynamicIPv4 *int32 `json:"dynamicIPv4,omitempty"`

	// DynamicIPv6 is how many IPv6 addresses the provider should pick.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=16
	DynamicIPv6 *int32 `json:"dynamicIPv6,omitempty"`

	// IPv4 are specific IPv4 addresses to request.
	// +kubebuilder:validation:Optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:items:MaxLength=15
	IPv4 []string `json:"ipv4,omitempty"`

	// IPv6 are specific IPv6 addresses to request.
	// +kubebuilder:validation:Optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:items:MaxLength=39
	IPv6 []string `json:"ipv6,omitempty"`
}

// NetworkProvisioning configures the source of the network segment.
type NetworkProvisioning struct {
	// Provider identifies the system that provisions the network. Well-known
	// values are "nam" (Network Administration Management backend, the default)
	// and "manual" (user-supplied config from the manual block), but this is an
	// open string so additional provisioners can be plugged in without a schema
	// change — each operator claims the NetworkNamespaces whose provider it owns.
	// Mirrors the open-string convention already used by ipAllocation.provider.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	// +kubebuilder:default=nam
	Provider string `json:"provider"`

	// Manual contains user-supplied network configuration.
	// Required when provider is "manual".
	// +kubebuilder:validation:Optional
	Manual *ManualProvisioningConfig `json:"manual,omitempty"`
}

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

	// Static contains optional configuration for static allocation within the
	// provisioned network. When omitted, defaults apply.
	// +kubebuilder:validation:Optional
	Static *StaticIPAllocationConfig `json:"static,omitempty"`

	// DHCP contains optional configuration overrides for DHCP-based allocation.
	// +kubebuilder:validation:Optional
	DHCP *DHCPAllocationConfig `json:"dhcp,omitempty"`
}

// StaticIPAllocationConfig configures static IP assignment within the
// provisioned network (status.ipv4Prefix). The subnet, gateway and VLAN
// come from networkProvisioning, not from here.
type StaticIPAllocationConfig struct {
	// IPv4RangeStart is the first allocatable address. Must be inside the
	// provisioned prefix. Defaults to the network address + 4.
	// +kubebuilder:validation:Optional
	IPv4RangeStart string `json:"ipv4RangeStart,omitempty"`

	// IPv4RangeEnd is the last allocatable address. Must be inside the
	// provisioned prefix. Defaults to the last usable address.
	// +kubebuilder:validation:Optional
	IPv4RangeEnd string `json:"ipv4RangeEnd,omitempty"`

	// DNS is a list of DNS server addresses for machines in this pool.
	// +kubebuilder:validation:Optional
	DNS []string `json:"dns,omitempty"`

	// TTLSeconds is the time-to-live for IP allocations in seconds.
	// After this period, unused allocations may be reclaimed by the operator.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:default=3600
	TTLSeconds int32 `json:"ttlSeconds,omitempty"`
}

// DHCPAllocationConfig provides optional configuration overrides when using
// DHCP-based IP allocation.
type DHCPAllocationConfig struct {
	// RequireClientClasses specifies Kea DHCP client classes that must be
	// matched for lease allocation.
	// +kubebuilder:validation:Optional
	RequireClientClasses []string `json:"requireClientClasses,omitempty"`
}

// NetworkNamespaceStatus defines the observed state of a NetworkNamespace.
type NetworkNamespaceStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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
	ProvisioningPhase ProvisioningPhase `json:"provisioningPhase,omitempty"`

	// Network fields populated by the provisioning layer.
	DataCenterID string `json:"datacenterId,omitempty"`
	NamespaceID  string `json:"namespaceId,omitempty"`
	IPv4Prefix   string `json:"ipv4Prefix,omitempty"`
	IPv6Prefix   string `json:"ipv6Prefix,omitempty"`
	VlanID       int    `json:"vlanId,omitempty"`

	// Egress reports the egress addresses assigned to the namespace
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=network
	Egress []EgressStatus `json:"egress,omitempty"`

	// IPAllocationSummary reports aggregate IP allocation state.
	// +kubebuilder:validation:Optional
	IPAllocationSummary *IPAllocationSummary `json:"ipAllocationSummary,omitempty"`
}

// EgressStatus reports the egress addresses assigned on one network.
type EgressStatus struct {
	// Network is the egress network.
	// +kubebuilder:validation:Required
	Network string `json:"network"`

	// IPv4 are the assigned IPv4 egress addresses, in provider order.
	// +kubebuilder:validation:Optional
	// +listType=atomic
	IPv4 []string `json:"ipv4,omitempty"`

	// IPv6 are the assigned IPv6 egress addresses, in provider order.
	// +kubebuilder:validation:Optional
	// +listType=atomic
	IPv6 []string `json:"ipv6,omitempty"`
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

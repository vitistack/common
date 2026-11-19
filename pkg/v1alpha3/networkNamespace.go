package v1alpha3

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
// +kubebuilder:ac:generate=true
// +kubebuilder:resource:path=networknamespaces,scope=Namespaced,shortName=nn
// // +kubebuilder:storageversion
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
	// NetworkProvisioning defines how the network segment (IP prefix + VLAN) is
	// acquired. When not set, defaults to provider "nam" for backward compatibility.
	// +kubebuilder:validation:Optional
	NetworkProvisioning *NetworkProvisioning `json:"networkProvisioning,omitempty"`

	// IPAllocation defines how individual IP addresses are assigned within the
	// provisioned network. When not set, defaults to DHCP-based allocation.
	// +kubebuilder:validation:Optional
	IPAllocation *NetworkNamespaceIPAllocation `json:"ipAllocation,omitempty"`

	// Egress requests egress addresses for the namespace, one entry per egress
	// network. Mutable; providers reconcile changes. A provider that cannot
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=network
	// +kubebuilder:validation:MaxItems=8
	Egress []EgressRequest `json:"egress,omitempty"`
}

// EgressNetwork names an egress network. It is an open string so providers can
// define their own networks without a schema change. Well-known values are
// "inet" and "hnet".
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=32
// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`
type EgressNetwork string

// EgressRequest requests egress addresses on one network.
// +kubebuilder:validation:XValidation:rule="has(self.ipv4) || has(self.ipv6)",message="at least one of ipv4 or ipv6 must be set"
// +kubebuilder:validation:XValidation:rule="!has(self.ipv4) || !has(self.ipv4.fixedAddresses) || self.ipv4.fixedAddresses.all(a, isIP(a) && ip(a).family() == 4)",message="ipv4.fixedAddresses must contain IPv4 addresses"
// +kubebuilder:validation:XValidation:rule="!has(self.ipv6) || !has(self.ipv6.fixedAddresses) || self.ipv6.fixedAddresses.all(a, isIP(a) && ip(a).family() == 6)",message="ipv6.fixedAddresses must contain IPv6 addresses"
type EgressRequest struct {
	// Network is the egress network the addresses are requested on.
	// +kubebuilder:validation:Required
	Network EgressNetwork `json:"network"`

	// IPv4 requests IPv4 egress addresses.
	// +kubebuilder:validation:Optional
	IPv4 *EgressAddressRequest `json:"ipv4,omitempty"`

	// IPv6 requests IPv6 egress addresses.
	// +kubebuilder:validation:Optional
	IPv6 *EgressAddressRequest `json:"ipv6,omitempty"`
}

// EgressAddressRequest requests egress addresses of one IP family.
// Exactly one of dynamicAddresses or fixedAddresses must be set.
// +kubebuilder:validation:XValidation:rule="has(self.dynamicAddresses) != has(self.fixedAddresses)",message="exactly one of dynamicAddresses or fixedAddresses must be set"
type EgressAddressRequest struct {
	// DynamicAddresses is the number of egress addresses to allocate.
	// A pointer so that 0 differs from unset.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4
	DynamicAddresses *int32 `json:"dynamicAddresses,omitempty"`

	// FixedAddresses are specific egress addresses to use. They must lie within
	// the datacenter's egress ranges; the provider checks this and reports
	// violations through the EgressReady condition.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:items:MaxLength=45
	// +listType=set
	FixedAddresses []string `json:"fixedAddresses,omitempty"`
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
	// +kubebuilder:validation:Enum=Pending;Ready;Error
	ProvisioningPhase ProvisioningPhase `json:"provisioningPhase,omitempty"`

	// Network fields populated by the provisioning layer.
	DataCenterIdentifier string `json:"datacenterIdentifier,omitempty"`
	NamespaceID          string `json:"namespaceId,omitempty"`
	IPv4Prefix           string `json:"ipv4Prefix,omitempty"`
	IPv6Prefix           string `json:"ipv6Prefix,omitempty"`
	VlanID               int    `json:"vlanId,omitempty"`

	// Egress reports the egress addresses assigned to the namespace
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=network
	Egress []EgressStatus `json:"egress,omitempty"`

	// IPAllocationSummary reports aggregate IP allocation state. This is a
	// +kubebuilder:validation:Optional
	IPAllocationSummary *IPAllocationSummary `json:"ipAllocationSummary,omitempty"`
}

// EgressStatus reports the egress addresses assigned on one network.
type EgressStatus struct {
	// Network is the egress network.
	// +kubebuilder:validation:Required
	Network EgressNetwork `json:"network"`

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

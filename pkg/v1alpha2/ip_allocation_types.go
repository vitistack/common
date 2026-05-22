/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import "strings"

// IPAllocationType specifies the method used for IP address allocation
// within a NetworkNamespace.
// +kubebuilder:validation:Enum=dhcp;static
type IPAllocationType string

const (
	// IPAllocationTypeDHCP allocates IP addresses via DHCP.
	IPAllocationTypeDHCP IPAllocationType = "dhcp"

	// IPAllocationTypeStatic allocates IP addresses from a statically defined range.
	IPAllocationTypeStatic IPAllocationType = "static"
)

// Well-known provider names used in spec.ipAllocation.provider and
// NetworkConfiguration.spec.provider to identify which operator handles
// the resource.
const (
	// ProviderNameKea identifies the kea-operator as the handler.
	ProviderNameKea = "kea"

	// ProviderNameStaticIP identifies the static-ip-operator as the handler.
	ProviderNameStaticIP = "static-ip-operator"
)

// NetworkProvisioningType identifies the system that provisions network
// segments (IP prefixes + VLANs) for a NetworkNamespace.
type NetworkProvisioningType string

const (
	// NetworkProvisioningNAM provisions networks through NAM (Network Administration Management).
	NetworkProvisioningNAM NetworkProvisioningType = "nam"

	// NetworkProvisioningManual uses user-supplied network configuration.
	NetworkProvisioningManual NetworkProvisioningType = "manual"
)

// ProvisioningPhase represents the lifecycle state of network provisioning.
type ProvisioningPhase string

const (
	ProvisioningPhasePending ProvisioningPhase = "Pending"
	ProvisioningPhaseReady   ProvisioningPhase = "Ready"
	ProvisioningPhaseError   ProvisioningPhase = "Error"
)

// IPAllocationPhase represents the lifecycle of an individual IP allocation.
type IPAllocationPhase string

const (
	IPAllocationPhasePending   IPAllocationPhase = "Pending"
	IPAllocationPhaseAllocated IPAllocationPhase = "Allocated"
	IPAllocationPhaseReleased  IPAllocationPhase = "Released"
	IPAllocationPhaseError     IPAllocationPhase = "Error"
)

// NormalizeProvider returns the lowercase, trimmed value of a provider string.
func NormalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// MatchesProvider checks whether the given raw provider string matches the
// expected operator name (case-insensitive, trimmed). Returns false if the
// provider is empty/whitespace.
func MatchesProvider(raw, expected string) bool {
	p := NormalizeProvider(raw)
	if p == "" {
		return false
	}
	return p == strings.ToLower(expected)
}

// IsProviderSet returns true if the provider string contains a non-empty,
// non-whitespace value.
func IsProviderSet(provider string) bool {
	return NormalizeProvider(provider) != ""
}

// IsValid returns true if the type is a known IP allocation type.
func (t IPAllocationType) IsValid() bool {
	switch t {
	case IPAllocationTypeDHCP, IPAllocationTypeStatic:
		return true
	default:
		return false
	}
}

// String returns the string representation of the type.
func (t IPAllocationType) String() string {
	return string(t)
}

// StaticIPAllocationConfig defines the IP address pool for static allocation.
type StaticIPAllocationConfig struct {
	// IPv4CIDR is the subnet in CIDR notation (e.g. "10.0.1.0/24").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`
	IPv4CIDR string `json:"ipv4CIDR"`

	// IPv4Gateway is the default gateway address for the subnet (e.g. "10.0.1.1").
	// +kubebuilder:validation:Required
	IPv4Gateway string `json:"ipv4Gateway"`

	// IPv4RangeStart is the first allocatable IP address in the range.
	// If not set, defaults to the second usable address in the CIDR.
	// +kubebuilder:validation:Optional
	IPv4RangeStart string `json:"ipv4RangeStart,omitempty"`

	// IPv4RangeEnd is the last allocatable IP address in the range.
	// If not set, defaults to the last usable address in the CIDR.
	// +kubebuilder:validation:Optional
	IPv4RangeEnd string `json:"ipv4RangeEnd,omitempty"`

	// VlanID is the VLAN ID for the subnet. When set, the kubevirt-operator
	// creates a NetworkAttachmentDefinition with this VLAN tag.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	VlanID int `json:"vlanId,omitempty"`

	// DNS is a list of DNS server addresses to use for machines in this pool.
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

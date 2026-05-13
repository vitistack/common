package v1alpha2

import (
	v1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
)

// ConvertNetworkNamespaceFromV1alpha1 converts a v1alpha1 NetworkNamespace to v1alpha2.
// This is the "spoke → hub" direction.
func ConvertNetworkNamespaceFromV1alpha1(src *v1alpha1.NetworkNamespace) *NetworkNamespace {
	dst := &NetworkNamespace{}
	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	dst.TypeMeta = src.TypeMeta
	dst.TypeMeta.APIVersion = GroupVersion.String()

	// --- Spec conversion ---
	dst.Spec.DatacenterIdentifier = src.Spec.DatacenterIdentifier
	dst.Spec.SupervisorIdentifier = src.Spec.SupervisorIdentifier

	// Infer networkProvisioning from v1alpha1 ipAllocation
	if src.Spec.IPAllocation != nil &&
		src.Spec.IPAllocation.Type == v1alpha1.IPAllocationTypeStatic &&
		src.Spec.IPAllocation.Static != nil {
		// Static with inline network config → manual provisioning + static allocation
		dst.Spec.NetworkProvisioning = &NetworkProvisioning{
			Provider: NetworkProvisioningManual,
			Manual: &ManualProvisioningConfig{
				IPv4CIDR:    src.Spec.IPAllocation.Static.IPv4CIDR,
				IPv4Gateway: src.Spec.IPAllocation.Static.IPv4Gateway,
				VlanID:      src.Spec.IPAllocation.Static.VlanID,
			},
		}
	} else {
		// DHCP or unset → NAM provisioning (original behavior)
		dst.Spec.NetworkProvisioning = &NetworkProvisioning{
			Provider: NetworkProvisioningNAM,
		}
	}

	// Convert ipAllocation
	if src.Spec.IPAllocation != nil {
		dst.Spec.IPAllocation = &NetworkNamespaceIPAllocation{
			Type:     IPAllocationType(src.Spec.IPAllocation.Type),
			Provider: src.Spec.IPAllocation.Provider,
		}
		if src.Spec.IPAllocation.Static != nil {
			dst.Spec.IPAllocation.Static = &StaticIPAllocationConfig{
				IPv4CIDR:       src.Spec.IPAllocation.Static.IPv4CIDR,
				IPv4Gateway:    src.Spec.IPAllocation.Static.IPv4Gateway,
				IPv4RangeStart: src.Spec.IPAllocation.Static.IPv4RangeStart,
				IPv4RangeEnd:   src.Spec.IPAllocation.Static.IPv4RangeEnd,
				VlanID:         src.Spec.IPAllocation.Static.VlanID,
				DNS:            src.Spec.IPAllocation.Static.DNS,
				TTLSeconds:     src.Spec.IPAllocation.Static.TTLSeconds,
			}
		}
		if src.Spec.IPAllocation.DHCP != nil {
			dst.Spec.IPAllocation.DHCP = &DHCPAllocationConfig{
				RequireClientClasses: src.Spec.IPAllocation.DHCP.RequireClientClasses,
			}
		}
	}

	// --- Status conversion ---
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Phase = src.Status.Phase
	dst.Status.Status = src.Status.Status
	dst.Status.Message = src.Status.Message
	dst.Status.Created = src.Status.Created
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.RetryCount = src.Status.RetryCount
	dst.Status.DataCenterIdentifier = src.Status.DataCenterIdentifier
	dst.Status.SupervisorIdentifier = src.Status.SupervisorIdentifier
	dst.Status.NamespaceID = src.Status.NamespaceID
	dst.Status.IPv4Prefix = src.Status.IPv4Prefix
	dst.Status.IPv6Prefix = src.Status.IPv6Prefix
	dst.Status.IPv4EgressIP = src.Status.IPv4EgressIP
	dst.Status.IPv6EgressIP = src.Status.IPv6EgressIP
	dst.Status.VlanID = src.Status.VlanID
	dst.Status.AssociatedKubernetesClusterIDs = src.Status.AssociatedKubernetesClusterIDs

	// Infer provisioningPhase from v1alpha1
	if src.Status.ProvisioningPhase != "" {
		dst.Status.ProvisioningPhase = ProvisioningPhase(src.Status.ProvisioningPhase)
	} else {
		// Backfill from phase for resources that predate the field
		switch src.Status.Phase {
		case "Ready":
			dst.Status.ProvisioningPhase = ProvisioningPhaseReady
		case "Error":
			dst.Status.ProvisioningPhase = ProvisioningPhaseError
		default:
			dst.Status.ProvisioningPhase = ProvisioningPhasePending
		}
	}

	// Convert IPAllocationStatus → IPAllocationSummary (drop allocatedIPs array)
	if src.Status.IPAllocationStatus != nil {
		dst.Status.IPAllocationSummary = &IPAllocationSummary{
			Type:           IPAllocationType(src.Status.IPAllocationStatus.Type),
			Provider:       src.Status.IPAllocationStatus.Provider,
			AllocatedCount: src.Status.IPAllocationStatus.AllocatedCount,
			AvailableCount: src.Status.IPAllocationStatus.AvailableCount,
			TotalCount:     src.Status.IPAllocationStatus.TotalCount,
		}
	}

	return dst
}

// ConvertNetworkNamespaceToV1alpha1 converts a v1alpha2 NetworkNamespace back to v1alpha1.
// This is the "hub → spoke" direction.
func ConvertNetworkNamespaceToV1alpha1(src *NetworkNamespace) *v1alpha1.NetworkNamespace {
	dst := &v1alpha1.NetworkNamespace{}
	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	dst.TypeMeta = src.TypeMeta
	dst.TypeMeta.APIVersion = v1alpha1.GroupVersion.String()

	// --- Spec conversion ---
	dst.Spec.DatacenterIdentifier = src.Spec.DatacenterIdentifier
	dst.Spec.SupervisorIdentifier = src.Spec.SupervisorIdentifier

	// Convert ipAllocation
	if src.Spec.IPAllocation != nil {
		dst.Spec.IPAllocation = &v1alpha1.NetworkNamespaceIPAllocation{
			Type:     v1alpha1.IPAllocationType(src.Spec.IPAllocation.Type),
			Provider: src.Spec.IPAllocation.Provider,
		}
		if src.Spec.IPAllocation.Static != nil {
			dst.Spec.IPAllocation.Static = &v1alpha1.StaticIPAllocationConfig{
				IPv4CIDR:       src.Spec.IPAllocation.Static.IPv4CIDR,
				IPv4Gateway:    src.Spec.IPAllocation.Static.IPv4Gateway,
				IPv4RangeStart: src.Spec.IPAllocation.Static.IPv4RangeStart,
				IPv4RangeEnd:   src.Spec.IPAllocation.Static.IPv4RangeEnd,
				VlanID:         src.Spec.IPAllocation.Static.VlanID,
				DNS:            src.Spec.IPAllocation.Static.DNS,
				TTLSeconds:     src.Spec.IPAllocation.Static.TTLSeconds,
			}
		}
		if src.Spec.IPAllocation.DHCP != nil {
			dst.Spec.IPAllocation.DHCP = &v1alpha1.DHCPAllocationConfig{
				RequireClientClasses: src.Spec.IPAllocation.DHCP.RequireClientClasses,
			}
		}
	} else if src.Spec.NetworkProvisioning != nil &&
		src.Spec.NetworkProvisioning.Provider == NetworkProvisioningManual &&
		src.Spec.NetworkProvisioning.Manual != nil {
		// When v1alpha2 has manual provisioning but no explicit ipAllocation,
		// we need to reconstruct the v1alpha1 static config from the manual block
		// so v1alpha1 consumers can still operate.
		dst.Spec.IPAllocation = &v1alpha1.NetworkNamespaceIPAllocation{
			Type:     v1alpha1.IPAllocationTypeStatic,
			Provider: v1alpha1.ProviderNameStaticIP,
			Static: &v1alpha1.StaticIPAllocationConfig{
				IPv4CIDR:    src.Spec.NetworkProvisioning.Manual.IPv4CIDR,
				IPv4Gateway: src.Spec.NetworkProvisioning.Manual.IPv4Gateway,
				VlanID:      src.Spec.NetworkProvisioning.Manual.VlanID,
			},
		}
	}

	// --- Status conversion ---
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Phase = src.Status.Phase
	dst.Status.Status = src.Status.Status
	dst.Status.Message = src.Status.Message
	dst.Status.Created = src.Status.Created
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.RetryCount = src.Status.RetryCount
	dst.Status.DataCenterIdentifier = src.Status.DataCenterIdentifier
	dst.Status.SupervisorIdentifier = src.Status.SupervisorIdentifier
	dst.Status.NamespaceID = src.Status.NamespaceID
	dst.Status.IPv4Prefix = src.Status.IPv4Prefix
	dst.Status.IPv6Prefix = src.Status.IPv6Prefix
	dst.Status.IPv4EgressIP = src.Status.IPv4EgressIP
	dst.Status.IPv6EgressIP = src.Status.IPv6EgressIP
	dst.Status.VlanID = src.Status.VlanID
	dst.Status.AssociatedKubernetesClusterIDs = src.Status.AssociatedKubernetesClusterIDs
	dst.Status.ProvisioningPhase = string(src.Status.ProvisioningPhase)

	// Convert IPAllocationSummary → IPAllocationStatus (no allocatedIPs to fill)
	if src.Status.IPAllocationSummary != nil {
		dst.Status.IPAllocationStatus = &v1alpha1.NetworkNamespaceIPAllocationStatus{
			Type:           v1alpha1.IPAllocationType(src.Status.IPAllocationSummary.Type),
			Provider:       src.Status.IPAllocationSummary.Provider,
			AllocatedCount: src.Status.IPAllocationSummary.AllocatedCount,
			AvailableCount: src.Status.IPAllocationSummary.AvailableCount,
			TotalCount:     src.Status.IPAllocationSummary.TotalCount,
		}
	}

	return dst
}

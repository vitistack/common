package v1alpha2

import (
	"fmt"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/vitistack/common/internal/nnconversion"
	v1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
)

// hubData holds v1alpha3 fields that v1alpha2 cannot represent. It is stashed
// on v1alpha2 objects in nnconversion.HubDataAnnotation.
// +kubebuilder:object:generate=false
type hubData struct {
	Egress       []v1alpha3.EgressRequest `json:"egress,omitempty"`
	StatusEgress []v1alpha3.EgressStatus  `json:"statusEgress,omitempty"`
}

// spokeData holds v1alpha2 fields that v1alpha3 cannot represent. It is
// stashed on v1alpha3 objects in nnconversion.V1alpha2DataAnnotation.
// +kubebuilder:object:generate=false
type spokeData struct {
	// Static is ipAllocation.static's network when it differs from
	// networkProvisioning.manual.
	Static                         *nnconversion.LegacyNet `json:"static,omitempty"`
	AssociatedKubernetesClusterIDs []string                `json:"associatedKubernetesClusterIds,omitempty"`
}

var _ conversion.Convertible = &NetworkNamespace{}

// ConvertTo converts this v1alpha2 NetworkNamespace to the v1alpha3 hub.
func (src *NetworkNamespace) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1alpha3.NetworkNamespace)
	if !ok {
		return fmt.Errorf("unsupported hub type %T", dstRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	var hub hubData
	if err := nnconversion.Take(&dst.ObjectMeta, nnconversion.HubDataAnnotation, &hub); err != nil {
		return err
	}

	dst.Spec = v1alpha3.NetworkNamespaceSpec{
		NetworkProvisioning: provisioningToHub(src.Spec.NetworkProvisioning),
		IPAllocation:        ipAllocationToHub(src.Spec.IPAllocation),
		Egress:              hub.Egress,
	}

	s := src.Status
	dst.Status = v1alpha3.NetworkNamespaceStatus{
		Conditions:         slices.Clone(s.Conditions),
		Phase:              s.Phase,
		Status:             s.Status,
		Message:            s.Message,
		Created:            s.Created,
		ObservedGeneration: s.ObservedGeneration,
		RetryCount:         s.RetryCount,
		ProvisioningPhase:  v1alpha3.ProvisioningPhase(s.ProvisioningPhase),
		DataCenterID:       s.DataCenterIdentifier,
		NamespaceID:        s.NamespaceID,
		IPv4Prefix:         s.IPv4Prefix,
		IPv6Prefix:         s.IPv6Prefix,
		VlanID:             s.VlanID,
		Egress:             nnconversion.MergeLegacyEgress(hub.StatusEgress, s.IPv4EgressIP, s.IPv6EgressIP),
	}
	if sum := s.IPAllocationSummary; sum != nil {
		dst.Status.IPAllocationSummary = &v1alpha3.IPAllocationSummary{
			Type:           v1alpha3.IPAllocationType(sum.Type),
			Provider:       sum.Provider,
			AllocatedCount: sum.AllocatedCount,
			AvailableCount: sum.AvailableCount,
			TotalCount:     sum.TotalCount,
		}
	}

	spoke := spokeData{AssociatedKubernetesClusterIDs: slices.Clone(s.AssociatedKubernetesClusterIDs)}
	if a := src.Spec.IPAllocation; a != nil && a.Static != nil {
		net := nnconversion.LegacyNet{IPv4CIDR: a.Static.IPv4CIDR, IPv4Gateway: a.Static.IPv4Gateway, VlanID: a.Static.VlanID}
		if net != nnconversion.ManualNet(dst.Spec.NetworkProvisioning) {
			spoke.Static = &net
		}
	}
	return nnconversion.Put(&dst.ObjectMeta, nnconversion.V1alpha2DataAnnotation, spoke)
}

// ConvertFrom converts the v1alpha3 hub to this v1alpha2 NetworkNamespace.
func (dst *NetworkNamespace) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha3.NetworkNamespace)
	if !ok {
		return fmt.Errorf("unsupported hub type %T", srcRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	var spoke spokeData
	if err := nnconversion.Take(&dst.ObjectMeta, nnconversion.V1alpha2DataAnnotation, &spoke); err != nil {
		return err
	}

	net := nnconversion.ManualNet(src.Spec.NetworkProvisioning)
	if spoke.Static != nil {
		net = *spoke.Static
	}
	dst.Spec = NetworkNamespaceSpec{
		NetworkProvisioning: provisioningFromHub(src.Spec.NetworkProvisioning),
		IPAllocation:        ipAllocationFromHub(src.Spec.IPAllocation, net),
	}

	s := src.Status
	ipv4Egress, ipv6Egress := nnconversion.LegacyEgress(s.Egress)
	dst.Status = NetworkNamespaceStatus{
		Conditions:                     slices.Clone(s.Conditions),
		Phase:                          s.Phase,
		Status:                         s.Status,
		Message:                        s.Message,
		Created:                        s.Created,
		ObservedGeneration:             s.ObservedGeneration,
		RetryCount:                     s.RetryCount,
		ProvisioningPhase:              ProvisioningPhase(s.ProvisioningPhase),
		DataCenterIdentifier:           s.DataCenterID,
		NamespaceID:                    s.NamespaceID,
		IPv4Prefix:                     s.IPv4Prefix,
		IPv6Prefix:                     s.IPv6Prefix,
		IPv4EgressIP:                   ipv4Egress,
		IPv6EgressIP:                   ipv6Egress,
		VlanID:                         s.VlanID,
		AssociatedKubernetesClusterIDs: spoke.AssociatedKubernetesClusterIDs,
	}
	if sum := s.IPAllocationSummary; sum != nil {
		dst.Status.IPAllocationSummary = &IPAllocationSummary{
			Type:           IPAllocationType(sum.Type),
			Provider:       sum.Provider,
			AllocatedCount: sum.AllocatedCount,
			AvailableCount: sum.AvailableCount,
			TotalCount:     sum.TotalCount,
		}
	}

	hub := hubData{Egress: nnconversion.DeepCopyEgressRequests(src.Spec.Egress)}
	if nnconversion.EgressNeedsStash(s.Egress) {
		hub.StatusEgress = nnconversion.DeepCopyEgressStatuses(s.Egress)
	}
	return nnconversion.Put(&dst.ObjectMeta, nnconversion.HubDataAnnotation, hub)
}

func provisioningToHub(p *NetworkProvisioning) *v1alpha3.NetworkProvisioning {
	if p == nil {
		return nil
	}
	out := &v1alpha3.NetworkProvisioning{Provider: string(p.Provider)}
	if m := p.Manual; m != nil {
		out.Manual = &v1alpha3.ManualProvisioningConfig{
			IPv4CIDR:    m.IPv4CIDR,
			IPv4Gateway: m.IPv4Gateway,
			IPv6CIDR:    m.IPv6CIDR,
			VlanID:      m.VlanID,
		}
	}
	return out
}

func provisioningFromHub(p *v1alpha3.NetworkProvisioning) *NetworkProvisioning {
	if p == nil {
		return nil
	}
	out := &NetworkProvisioning{Provider: NetworkProvisioningType(p.Provider)}
	if m := p.Manual; m != nil {
		out.Manual = &ManualProvisioningConfig{
			IPv4CIDR:    m.IPv4CIDR,
			IPv4Gateway: m.IPv4Gateway,
			IPv6CIDR:    m.IPv6CIDR,
			VlanID:      m.VlanID,
		}
	}
	return out
}

func ipAllocationToHub(a *NetworkNamespaceIPAllocation) *v1alpha3.NetworkNamespaceIPAllocation {
	if a == nil {
		return nil
	}
	out := &v1alpha3.NetworkNamespaceIPAllocation{
		Type:     v1alpha3.IPAllocationType(a.Type),
		Provider: a.Provider,
	}
	if s := a.Static; s != nil {
		out.Static = &v1alpha3.StaticIPAllocationConfig{
			IPv4RangeStart: s.IPv4RangeStart,
			IPv4RangeEnd:   s.IPv4RangeEnd,
			DNS:            slices.Clone(s.DNS),
			TTLSeconds:     s.TTLSeconds,
		}
	}
	if d := a.DHCP; d != nil {
		out.DHCP = &v1alpha3.DHCPAllocationConfig{RequireClientClasses: slices.Clone(d.RequireClientClasses)}
	}
	return out
}

func ipAllocationFromHub(a *v1alpha3.NetworkNamespaceIPAllocation, net nnconversion.LegacyNet) *NetworkNamespaceIPAllocation {
	if a == nil {
		return nil
	}
	out := &NetworkNamespaceIPAllocation{
		Type:     IPAllocationType(a.Type),
		Provider: a.Provider,
	}
	if s := a.Static; s != nil {
		out.Static = &StaticIPAllocationConfig{
			IPv4CIDR:       net.IPv4CIDR,
			IPv4Gateway:    net.IPv4Gateway,
			IPv4RangeStart: s.IPv4RangeStart,
			IPv4RangeEnd:   s.IPv4RangeEnd,
			VlanID:         net.VlanID,
			DNS:            slices.Clone(s.DNS),
			TTLSeconds:     s.TTLSeconds,
		}
	}
	if d := a.DHCP; d != nil {
		out.DHCP = &DHCPAllocationConfig{RequireClientClasses: slices.Clone(d.RequireClientClasses)}
	}
	return out
}

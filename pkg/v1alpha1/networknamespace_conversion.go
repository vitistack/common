package v1alpha1

import (
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/vitistack/common/internal/nnconversion"
	v1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
)

// Annotations that carry v1alpha1's spec.datacenterIdentifier and
// spec.supervisorIdentifier on later versions, which dropped those fields.
const (
	datacenterIdentifierAnnotation = "datacenterIdentifier"
	supervisorIdentifierAnnotation = "supervisorIdentifier"
)

// hubData holds v1alpha3 fields that v1alpha1 cannot represent. It is stashed
// on v1alpha1 objects in nnconversion.HubDataAnnotation.
// +kubebuilder:object:generate=false
type hubData struct {
	Egress       []v1alpha3.EgressRequest `json:"egress,omitempty"`
	StatusEgress []v1alpha3.EgressStatus  `json:"statusEgress,omitempty"`

	// NetworkProvisioning is set when inferring it from ipAllocation would
	// not give back the hub's value.
	NetworkProvisioning *provisioningStash `json:"networkProvisioning,omitempty"`
	// IPAllocationUnset records that the hub had no ipAllocation and the
	// static one v1alpha1 shows was synthesised from manual provisioning.
	IPAllocationUnset bool `json:"ipAllocationUnset,omitempty"`
	// ProvisioningPhaseUnset records that the hub had no provisioningPhase,
	// so it must not be backfilled from phase.
	ProvisioningPhaseUnset bool `json:"provisioningPhaseUnset,omitempty"`
}

// provisioningStash wraps the hub's networkProvisioning so that nil survives.
// +kubebuilder:object:generate=false
type provisioningStash struct {
	Value *v1alpha3.NetworkProvisioning `json:"value"`
}

// spokeData holds v1alpha1 fields that v1alpha3 cannot represent. It is
// stashed on v1alpha3 objects in nnconversion.V1alpha1DataAnnotation.
// +kubebuilder:object:generate=false
type spokeData struct {
	// Static is ipAllocation.static's network when provisioning is not
	// manual, so it cannot live in networkProvisioning.manual.
	Static                         *nnconversion.LegacyNet `json:"static,omitempty"`
	AssociatedKubernetesClusterIDs []string                `json:"associatedKubernetesClusterIds,omitempty"`
	AllocatedIPs                   []AllocatedIPEntry      `json:"allocatedIPs,omitempty"`
	// StatusSupervisorIdentifier is set when status.supervisorIdentifier
	// differs from the supervisorIdentifier annotation.
	StatusSupervisorIdentifier *string `json:"statusSupervisorIdentifier,omitempty"`
}

var _ conversion.Convertible = &NetworkNamespace{}

// ConvertTo converts this v1alpha1 NetworkNamespace to the v1alpha3 hub.
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
	setAnnotation(dst, datacenterIdentifierAnnotation, src.Spec.DatacenterIdentifier)
	setAnnotation(dst, supervisorIdentifierAnnotation, src.Spec.SupervisorIdentifier)

	prov := inferProvisioning(src.Spec.IPAllocation)
	if hub.NetworkProvisioning != nil {
		prov = hub.NetworkProvisioning.Value
		// v1alpha1 edits the manual network through ipAllocation.static.
		if isManual(prov) && hasStaticConfig(src.Spec.IPAllocation) {
			st := src.Spec.IPAllocation.Static
			prov.Manual.IPv4CIDR = st.IPv4CIDR
			prov.Manual.IPv4Gateway = st.IPv4Gateway
			prov.Manual.VlanID = st.VlanID
		}
	}

	ipAllocation := ipAllocationToHub(src.Spec.IPAllocation)
	if hub.IPAllocationUnset && equality.Semantic.DeepEqual(src.Spec.IPAllocation, synthesizeIPAllocation(prov)) {
		ipAllocation = nil
	}

	dst.Spec = v1alpha3.NetworkNamespaceSpec{
		NetworkProvisioning: prov,
		IPAllocation:        ipAllocation,
		Egress:              hub.Egress,
	}

	s := src.Status
	dst.Status = v1alpha3.NetworkNamespaceStatus{
		Conditions:           slices.Clone(s.Conditions),
		Phase:                s.Phase,
		Status:               s.Status,
		Message:              s.Message,
		Created:              s.Created,
		ObservedGeneration:   s.ObservedGeneration,
		RetryCount:           s.RetryCount,
		ProvisioningPhase:    provisioningPhaseToHub(s.ProvisioningPhase, s.Phase, hub.ProvisioningPhaseUnset),
		DataCenterIdentifier: s.DataCenterIdentifier,
		NamespaceID:          s.NamespaceID,
		IPv4Prefix:           s.IPv4Prefix,
		IPv6Prefix:           s.IPv6Prefix,
		VlanID:               s.VlanID,
		Egress:               nnconversion.MergeLegacyEgress(hub.StatusEgress, s.IPv4EgressIP, s.IPv6EgressIP),
	}

	spoke := spokeData{AssociatedKubernetesClusterIDs: slices.Clone(s.AssociatedKubernetesClusterIDs)}
	if st := s.IPAllocationStatus; st != nil {
		dst.Status.IPAllocationSummary = &v1alpha3.IPAllocationSummary{
			Type:           v1alpha3.IPAllocationType(st.Type),
			Provider:       st.Provider,
			AllocatedCount: st.AllocatedCount,
			AvailableCount: st.AvailableCount,
			TotalCount:     st.TotalCount,
		}
		spoke.AllocatedIPs = slices.Clone(st.AllocatedIPs)
	}
	if !isManual(prov) && hasStaticConfig(src.Spec.IPAllocation) {
		st := src.Spec.IPAllocation.Static
		if net := (nnconversion.LegacyNet{IPv4CIDR: st.IPv4CIDR, IPv4Gateway: st.IPv4Gateway, VlanID: st.VlanID}); net != (nnconversion.LegacyNet{}) {
			spoke.Static = &net
		}
	}
	if s.SupervisorIdentifier != dst.Annotations[supervisorIdentifierAnnotation] {
		spoke.StatusSupervisorIdentifier = &s.SupervisorIdentifier
	}
	return nnconversion.Put(&dst.ObjectMeta, nnconversion.V1alpha1DataAnnotation, spoke)
}

// ConvertFrom converts the v1alpha3 hub to this v1alpha1 NetworkNamespace.
func (dst *NetworkNamespace) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha3.NetworkNamespace)
	if !ok {
		return fmt.Errorf("unsupported hub type %T", srcRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	var spoke spokeData
	if err := nnconversion.Take(&dst.ObjectMeta, nnconversion.V1alpha1DataAnnotation, &spoke); err != nil {
		return err
	}

	hub := hubData{Egress: nnconversion.DeepCopyEgressRequests(src.Spec.Egress)}

	prov := src.Spec.NetworkProvisioning
	net := nnconversion.ManualNet(prov)
	if !isManual(prov) && spoke.Static != nil {
		net = *spoke.Static
	}
	ipAllocation := ipAllocationFromHub(src.Spec.IPAllocation, net)
	if src.Spec.IPAllocation == nil {
		// v1alpha1 consumers expect the manual network in a static block.
		if ipAllocation = synthesizeIPAllocation(prov); ipAllocation != nil {
			hub.IPAllocationUnset = true
		}
	}
	if !equality.Semantic.DeepEqual(inferProvisioning(ipAllocation), prov) {
		hub.NetworkProvisioning = &provisioningStash{Value: prov.DeepCopy()}
	}

	dst.Spec = NetworkNamespaceSpec{
		DatacenterIdentifier: src.Annotations[datacenterIdentifierAnnotation],
		SupervisorIdentifier: src.Annotations[supervisorIdentifierAnnotation],
		IPAllocation:         ipAllocation,
	}

	s := src.Status
	ipv4Egress, ipv6Egress := nnconversion.LegacyEgress(s.Egress)
	supervisor := src.Annotations[supervisorIdentifierAnnotation]
	if spoke.StatusSupervisorIdentifier != nil {
		supervisor = *spoke.StatusSupervisorIdentifier
	}
	dst.Status = NetworkNamespaceStatus{
		Conditions:                     slices.Clone(s.Conditions),
		Phase:                          s.Phase,
		Status:                         s.Status,
		Message:                        s.Message,
		Created:                        s.Created,
		ObservedGeneration:             s.ObservedGeneration,
		RetryCount:                     s.RetryCount,
		ProvisioningPhase:              string(s.ProvisioningPhase),
		DataCenterIdentifier:           s.DataCenterIdentifier,
		SupervisorIdentifier:           supervisor,
		NamespaceID:                    s.NamespaceID,
		IPv4Prefix:                     s.IPv4Prefix,
		IPv6Prefix:                     s.IPv6Prefix,
		IPv4EgressIP:                   ipv4Egress,
		IPv6EgressIP:                   ipv6Egress,
		VlanID:                         s.VlanID,
		AssociatedKubernetesClusterIDs: spoke.AssociatedKubernetesClusterIDs,
	}
	if sum := s.IPAllocationSummary; sum != nil {
		dst.Status.IPAllocationStatus = &NetworkNamespaceIPAllocationStatus{
			Type:           IPAllocationType(sum.Type),
			Provider:       sum.Provider,
			AllocatedCount: sum.AllocatedCount,
			AvailableCount: sum.AvailableCount,
			TotalCount:     sum.TotalCount,
			AllocatedIPs:   spoke.AllocatedIPs,
		}
	}

	hub.ProvisioningPhaseUnset = s.ProvisioningPhase == "" && s.Phase != ""
	if nnconversion.EgressNeedsStash(s.Egress) {
		hub.StatusEgress = nnconversion.DeepCopyEgressStatuses(s.Egress)
	}
	return nnconversion.Put(&dst.ObjectMeta, nnconversion.HubDataAnnotation, hub)
}

func setAnnotation(obj *v1alpha3.NetworkNamespace, key, value string) {
	if value == "" {
		return
	}
	if obj.Annotations == nil {
		obj.Annotations = map[string]string{}
	}
	obj.Annotations[key] = value
}

func isManual(p *v1alpha3.NetworkProvisioning) bool {
	return p != nil && p.Provider == nnconversion.ProviderManual && p.Manual != nil
}

func hasStaticConfig(a *NetworkNamespaceIPAllocation) bool {
	return a != nil && a.Type == IPAllocationTypeStatic && a.Static != nil
}

// inferProvisioning derives networkProvisioning for a v1alpha1 object that
// has none: static allocation with a network means manual, otherwise NAM.
func inferProvisioning(a *NetworkNamespaceIPAllocation) *v1alpha3.NetworkProvisioning {
	if !hasStaticConfig(a) {
		return &v1alpha3.NetworkProvisioning{Provider: nnconversion.ProviderNAM}
	}
	return &v1alpha3.NetworkProvisioning{
		Provider: nnconversion.ProviderManual,
		Manual: &v1alpha3.ManualProvisioningConfig{
			IPv4CIDR:    a.Static.IPv4CIDR,
			IPv4Gateway: a.Static.IPv4Gateway,
			VlanID:      a.Static.VlanID,
		},
	}
}

// synthesizeIPAllocation builds the static ipAllocation v1alpha1 shows for
// manual provisioning without an explicit ipAllocation, or nil otherwise.
func synthesizeIPAllocation(p *v1alpha3.NetworkProvisioning) *NetworkNamespaceIPAllocation {
	if !isManual(p) {
		return nil
	}
	return &NetworkNamespaceIPAllocation{
		Type:     IPAllocationTypeStatic,
		Provider: ProviderNameStaticIP,
		Static: &StaticIPAllocationConfig{
			IPv4CIDR:    p.Manual.IPv4CIDR,
			IPv4Gateway: p.Manual.IPv4Gateway,
			VlanID:      p.Manual.VlanID,
		},
	}
}

// provisioningPhaseToHub backfills provisioningPhase from phase for objects
// written by clients that predate the field.
func provisioningPhaseToHub(provisioningPhase, phase string, keepUnset bool) v1alpha3.ProvisioningPhase {
	switch {
	case provisioningPhase != "":
		return v1alpha3.ProvisioningPhase(provisioningPhase)
	case phase == "" || keepUnset:
		return ""
	case phase == string(v1alpha3.ProvisioningPhaseReady):
		return v1alpha3.ProvisioningPhaseReady
	case phase == string(v1alpha3.ProvisioningPhaseError):
		return v1alpha3.ProvisioningPhaseError
	default:
		return v1alpha3.ProvisioningPhasePending
	}
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

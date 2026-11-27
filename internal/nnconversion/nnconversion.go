// Package nnconversion holds helpers shared by the NetworkNamespace spoke
// conversions (v1alpha1, v1alpha2 <-> v1alpha3 hub).
//
// Conversions must round-trip without loss. Fields a version cannot represent
// are stashed as JSON in an annotation on the converted object and restored
// when it is converted back.
package nnconversion

import (
	"encoding/json"
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
)

const (
	// HubDataAnnotation holds hub (v1alpha3) fields on a spoke object.
	HubDataAnnotation = "vitistack.io/conversion-data-v1alpha3"
	// V1alpha2DataAnnotation holds v1alpha2-only fields on a hub object.
	V1alpha2DataAnnotation = "vitistack.io/conversion-data-v1alpha2"
	// V1alpha1DataAnnotation holds v1alpha1-only fields on a hub object.
	V1alpha1DataAnnotation = "vitistack.io/conversion-data-v1alpha1"

	// LegacyEgressNetwork is the v1alpha3 egress network that the single
	// ipv4EgressIp/ipv6EgressIp of v1alpha1 and v1alpha2 map to.
	LegacyEgressNetwork = "hnet"

	ProviderNAM    = "nam"
	ProviderManual = "manual"
)

// Put stores data as JSON in annotation key. Zero-valued data removes the
// annotation instead, so objects without lossy fields stay clean.
func Put[T any](meta *metav1.ObjectMeta, key string, data T) error {
	var zero T
	if equality.Semantic.DeepEqual(data, zero) {
		deleteAnnotation(meta, key)
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshalling %s: %w", key, err)
	}
	if meta.Annotations == nil {
		meta.Annotations = map[string]string{}
	}
	meta.Annotations[key] = string(raw)
	return nil
}

// Take decodes annotation key into data and removes the annotation. It
// leaves data untouched when the annotation is absent.
func Take[T any](meta *metav1.ObjectMeta, key string, data *T) error {
	raw, ok := meta.Annotations[key]
	if !ok {
		return nil
	}
	deleteAnnotation(meta, key)
	if err := json.Unmarshal([]byte(raw), data); err != nil {
		return fmt.Errorf("unmarshalling %s: %w", key, err)
	}
	return nil
}

func deleteAnnotation(meta *metav1.ObjectMeta, key string) {
	delete(meta.Annotations, key)
	if len(meta.Annotations) == 0 {
		meta.Annotations = nil
	}
}

// LegacyNet is the subnet, gateway and VLAN that v1alpha1 and v1alpha2 carry
// in ipAllocation.static, and v1alpha3 only in networkProvisioning.manual.
type LegacyNet struct {
	IPv4CIDR    string `json:"ipv4CIDR,omitempty"`
	IPv4Gateway string `json:"ipv4Gateway,omitempty"`
	VlanID      int    `json:"vlanId,omitempty"`
}

// ManualNet is the LegacyNet a spoke's static block gets when nothing was
// stashed: the manual provisioning network, or zero for other providers.
func ManualNet(p *v1alpha3.NetworkProvisioning) LegacyNet {
	if p == nil || p.Provider != ProviderManual || p.Manual == nil {
		return LegacyNet{}
	}
	return LegacyNet{
		IPv4CIDR:    p.Manual.IPv4CIDR,
		IPv4Gateway: p.Manual.IPv4Gateway,
		VlanID:      p.Manual.VlanID,
	}
}

// LegacyEgress returns the single IPv4 and IPv6 egress addresses a spoke
// shows: the first of each on LegacyEgressNetwork.
func LegacyEgress(egress []v1alpha3.EgressStatus) (ipv4, ipv6 string) {
	for _, e := range egress {
		if e.Network != LegacyEgressNetwork {
			continue
		}
		if len(e.IPv4) > 0 {
			ipv4 = e.IPv4[0]
		}
		if len(e.IPv6) > 0 {
			ipv6 = e.IPv6[0]
		}
		return ipv4, ipv6
	}
	return "", ""
}

// MergeLegacyEgress rebuilds the hub's status.egress from the stashed hub
// value and a spoke's single egress addresses. Stashed addresses are kept
// unless the spoke changed the address it shows; then the spoke wins.
func MergeLegacyEgress(stashed []v1alpha3.EgressStatus, ipv4, ipv6 string) []v1alpha3.EgressStatus {
	out := make([]v1alpha3.EgressStatus, 0, len(stashed)+1)
	for _, e := range stashed {
		out = append(out, *e.DeepCopy())
	}

	i := slices.IndexFunc(out, func(e v1alpha3.EgressStatus) bool { return e.Network == LegacyEgressNetwork })
	if i < 0 {
		if ipv4 == "" && ipv6 == "" {
			return nilIfEmpty(out)
		}
		out = append(out, v1alpha3.EgressStatus{Network: LegacyEgressNetwork})
		i = len(out) - 1
	}

	e := &out[i]
	changed4, changed6 := first(e.IPv4) != ipv4, first(e.IPv6) != ipv6
	if !changed4 && !changed6 {
		return nilIfEmpty(out)
	}
	if changed4 {
		e.IPv4 = single(ipv4)
	}
	if changed6 {
		e.IPv6 = single(ipv6)
	}
	if len(e.IPv4) == 0 && len(e.IPv6) == 0 {
		out = slices.Delete(out, i, i+1)
	}
	return nilIfEmpty(out)
}

func first(ips []string) string {
	if len(ips) == 0 {
		return ""
	}
	return ips[0]
}

func single(ip string) []string {
	if ip == "" {
		return nil
	}
	return []string{ip}
}

// EgressNeedsStash reports whether egress holds more than a spoke's single
// addresses can rebuild.
func EgressNeedsStash(egress []v1alpha3.EgressStatus) bool {
	ipv4, ipv6 := LegacyEgress(egress)
	return !equality.Semantic.DeepEqual(egress, MergeLegacyEgress(nil, ipv4, ipv6))
}

func nilIfEmpty[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}

// DeepCopyEgressRequests deep-copies in, keeping nil as nil.
func DeepCopyEgressRequests(in []v1alpha3.EgressRequest) []v1alpha3.EgressRequest {
	if in == nil {
		return nil
	}
	out := make([]v1alpha3.EgressRequest, len(in))
	for i := range in {
		in[i].DeepCopyInto(&out[i])
	}
	return out
}

// DeepCopyEgressStatuses deep-copies in, keeping nil as nil.
func DeepCopyEgressStatuses(in []v1alpha3.EgressStatus) []v1alpha3.EgressStatus {
	if in == nil {
		return nil
	}
	out := make([]v1alpha3.EgressStatus, len(in))
	for i := range in {
		in[i].DeepCopyInto(&out[i])
	}
	return out
}

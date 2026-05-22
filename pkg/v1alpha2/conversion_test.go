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

import (
	"testing"

	v1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testNamespace  = "default"
	testDC         = "dc1"
	testSV         = "sv1"
	testIPv4Prefix = "10.0.1.0/24"
	testStaticCIDR = "10.0.2.0/24"
)

func TestConvertFromV1alpha1_NoIPAllocation(t *testing.T) {
	src := &v1alpha1.NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ns", Namespace: testNamespace},
		Spec: v1alpha1.NetworkNamespaceSpec{
			DatacenterIdentifier: testDC,
			SupervisorIdentifier: testSV,
		},
		Status: v1alpha1.NetworkNamespaceStatus{
			Phase:      string(ProvisioningPhaseReady),
			IPv4Prefix: testIPv4Prefix,
			VlanID:     100,
		},
	}

	dst := ConvertNetworkNamespaceFromV1alpha1(src)

	if dst.Spec.NetworkProvisioning == nil {
		t.Fatal("expected networkProvisioning to be set")
	}
	if dst.Spec.NetworkProvisioning.Provider != NetworkProvisioningNAM {
		t.Errorf("expected provider 'nam', got %q", dst.Spec.NetworkProvisioning.Provider)
	}
	if dst.Spec.IPAllocation != nil {
		t.Error("expected ipAllocation to be nil when no IPAllocation is set in v1alpha1")
	}
	if dst.Status.ProvisioningPhase != ProvisioningPhaseReady {
		t.Errorf("expected provisioningPhase 'Ready', got %q", dst.Status.ProvisioningPhase)
	}
	if dst.Status.IPv4Prefix != testIPv4Prefix {
		t.Errorf("expected ipv4Prefix preserved, got %q", dst.Status.IPv4Prefix)
	}
}

func TestConvertFromV1alpha1_DHCP(t *testing.T) {
	src := &v1alpha1.NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "dhcp-ns", Namespace: testNamespace},
		Spec: v1alpha1.NetworkNamespaceSpec{
			DatacenterIdentifier: testDC,
			SupervisorIdentifier: testSV,
			IPAllocation: &v1alpha1.NetworkNamespaceIPAllocation{
				Type:     v1alpha1.IPAllocationTypeDHCP,
				Provider: v1alpha1.ProviderNameKea,
				DHCP: &v1alpha1.DHCPAllocationConfig{
					RequireClientClasses: []string{"class1", "class2"},
				},
			},
		},
		Status: v1alpha1.NetworkNamespaceStatus{
			Phase:      string(ProvisioningPhaseReady),
			IPv4Prefix: testIPv4Prefix,
			VlanID:     100,
		},
	}

	dst := ConvertNetworkNamespaceFromV1alpha1(src)

	if dst.Spec.NetworkProvisioning == nil {
		t.Fatal("expected networkProvisioning to be set")
	}
	if dst.Spec.NetworkProvisioning.Provider != NetworkProvisioningNAM {
		t.Errorf("expected provider 'nam', got %q", dst.Spec.NetworkProvisioning.Provider)
	}
	if dst.Spec.IPAllocation == nil {
		t.Fatal("expected ipAllocation to be set for DHCP")
	}
	if dst.Spec.IPAllocation.Type != IPAllocationTypeDHCP {
		t.Errorf("expected ipAllocation.type 'dhcp', got %q", dst.Spec.IPAllocation.Type)
	}
	if dst.Spec.IPAllocation.DHCP == nil {
		t.Fatal("expected ipAllocation.dhcp to be set")
	}
	if len(dst.Spec.IPAllocation.DHCP.RequireClientClasses) != 2 {
		t.Errorf("expected 2 client classes, got %d", len(dst.Spec.IPAllocation.DHCP.RequireClientClasses))
	}
}

func TestConvertFromV1alpha1_Static(t *testing.T) {
	src := &v1alpha1.NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "static-ns", Namespace: testNamespace},
		Spec: v1alpha1.NetworkNamespaceSpec{
			DatacenterIdentifier: testDC,
			SupervisorIdentifier: testSV,
			IPAllocation: &v1alpha1.NetworkNamespaceIPAllocation{
				Type:     v1alpha1.IPAllocationTypeStatic,
				Provider: v1alpha1.ProviderNameStaticIP,
				Static: &v1alpha1.StaticIPAllocationConfig{
					IPv4CIDR:    testStaticCIDR,
					IPv4Gateway: "10.0.2.1",
					VlanID:      200,
					DNS:         []string{"8.8.8.8"},
					TTLSeconds:  3600,
				},
			},
		},
		Status: v1alpha1.NetworkNamespaceStatus{
			Phase:      string(ProvisioningPhaseReady),
			IPv4Prefix: testStaticCIDR,
			VlanID:     200,
			IPAllocationStatus: &v1alpha1.NetworkNamespaceIPAllocationStatus{
				Type:           v1alpha1.IPAllocationTypeStatic,
				Provider:       v1alpha1.ProviderNameStaticIP,
				TotalCount:     253,
				AllocatedCount: 5,
				AvailableCount: 248,
				AllocatedIPs: []v1alpha1.AllocatedIPEntry{
					{IP: "10.0.2.10", NetworkConfiguration: "nc-1"},
				},
			},
		},
	}

	dst := ConvertNetworkNamespaceFromV1alpha1(src)

	// Network provisioning should be manual
	if dst.Spec.NetworkProvisioning == nil {
		t.Fatal("expected networkProvisioning to be set")
	}
	if dst.Spec.NetworkProvisioning.Provider != NetworkProvisioningManual {
		t.Errorf("expected provider 'manual', got %q", dst.Spec.NetworkProvisioning.Provider)
	}
	if dst.Spec.NetworkProvisioning.Manual == nil {
		t.Fatal("expected manual config to be set")
	}
	if dst.Spec.NetworkProvisioning.Manual.IPv4CIDR != testStaticCIDR {
		t.Errorf("expected manual.ipv4CIDR %q, got %q", testStaticCIDR, dst.Spec.NetworkProvisioning.Manual.IPv4CIDR)
	}

	// IPAllocation should be preserved
	if dst.Spec.IPAllocation == nil {
		t.Fatal("expected ipAllocation to be set")
	}
	if dst.Spec.IPAllocation.Type != IPAllocationTypeStatic {
		t.Errorf("expected ipAllocation.type 'static', got %q", dst.Spec.IPAllocation.Type)
	}
	if dst.Spec.IPAllocation.Static == nil {
		t.Fatal("expected ipAllocation.static to be set")
	}

	// Summary should drop allocatedIPs
	if dst.Status.IPAllocationSummary == nil {
		t.Fatal("expected ipAllocationSummary to be set")
	}
	if dst.Status.IPAllocationSummary.TotalCount != 253 {
		t.Errorf("expected totalCount 253, got %d", dst.Status.IPAllocationSummary.TotalCount)
	}
}

func TestRoundTrip_NAM(t *testing.T) {
	original := &v1alpha1.NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "rt-ns", Namespace: testNamespace},
		Spec: v1alpha1.NetworkNamespaceSpec{
			DatacenterIdentifier: testDC,
			SupervisorIdentifier: testSV,
		},
		Status: v1alpha1.NetworkNamespaceStatus{
			Phase:      string(ProvisioningPhaseReady),
			IPv4Prefix: testIPv4Prefix,
			VlanID:     100,
		},
	}

	// v1alpha1 → v1alpha2 → v1alpha1
	v2 := ConvertNetworkNamespaceFromV1alpha1(original)
	roundtripped := ConvertNetworkNamespaceToV1alpha1(v2)

	if roundtripped.Spec.DatacenterIdentifier != original.Spec.DatacenterIdentifier {
		t.Errorf("datacenterIdentifier not preserved: %q vs %q", roundtripped.Spec.DatacenterIdentifier, original.Spec.DatacenterIdentifier)
	}
	if roundtripped.Status.IPv4Prefix != original.Status.IPv4Prefix {
		t.Errorf("ipv4Prefix not preserved: %q vs %q", roundtripped.Status.IPv4Prefix, original.Status.IPv4Prefix)
	}
	if roundtripped.Status.VlanID != original.Status.VlanID {
		t.Errorf("vlanId not preserved: %d vs %d", roundtripped.Status.VlanID, original.Status.VlanID)
	}
}

func TestRoundTrip_Static(t *testing.T) {
	original := &v1alpha1.NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "rt-static", Namespace: testNamespace},
		Spec: v1alpha1.NetworkNamespaceSpec{
			DatacenterIdentifier: testDC,
			SupervisorIdentifier: testSV,
			IPAllocation: &v1alpha1.NetworkNamespaceIPAllocation{
				Type:     v1alpha1.IPAllocationTypeStatic,
				Provider: v1alpha1.ProviderNameStaticIP,
				Static: &v1alpha1.StaticIPAllocationConfig{
					IPv4CIDR:       "10.0.3.0/24",
					IPv4Gateway:    "10.0.3.1",
					IPv4RangeStart: "10.0.3.10",
					IPv4RangeEnd:   "10.0.3.250",
					VlanID:         300,
					DNS:            []string{"1.1.1.1"},
					TTLSeconds:     7200,
				},
			},
		},
		Status: v1alpha1.NetworkNamespaceStatus{
			Phase: string(ProvisioningPhaseReady),
		},
	}

	v2 := ConvertNetworkNamespaceFromV1alpha1(original)
	roundtripped := ConvertNetworkNamespaceToV1alpha1(v2)

	if roundtripped.Spec.IPAllocation == nil {
		t.Fatal("expected ipAllocation to survive round-trip")
	}
	if roundtripped.Spec.IPAllocation.Static == nil {
		t.Fatal("expected static config to survive round-trip")
	}
	if roundtripped.Spec.IPAllocation.Static.IPv4CIDR != "10.0.3.0/24" {
		t.Errorf("ipv4CIDR not preserved: %q", roundtripped.Spec.IPAllocation.Static.IPv4CIDR)
	}
	if roundtripped.Spec.IPAllocation.Static.IPv4RangeStart != "10.0.3.10" {
		t.Errorf("ipv4RangeStart not preserved: %q", roundtripped.Spec.IPAllocation.Static.IPv4RangeStart)
	}
	if roundtripped.Spec.IPAllocation.Static.TTLSeconds != 7200 {
		t.Errorf("ttlSeconds not preserved: %d", roundtripped.Spec.IPAllocation.Static.TTLSeconds)
	}
}

func TestConvertToV1alpha1_ManualProvisioningOnly(t *testing.T) {
	// Test the branch where v1alpha2 has manual provisioning but no explicit IPAllocation
	src := &NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "manual-only", Namespace: testNamespace},
		Spec: NetworkNamespaceSpec{
			DatacenterIdentifier: testDC,
			SupervisorIdentifier: testSV,
			NetworkProvisioning: &NetworkProvisioning{
				Provider: NetworkProvisioningManual,
				Manual: &ManualProvisioningConfig{
					IPv4CIDR:    "10.0.4.0/24",
					IPv4Gateway: "10.0.4.1",
					VlanID:      400,
				},
			},
			// No IPAllocation set
		},
		Status: NetworkNamespaceStatus{
			ProvisioningPhase: ProvisioningPhaseReady,
		},
	}

	dst := ConvertNetworkNamespaceToV1alpha1(src)

	// Should reconstruct v1alpha1 static IPAllocation from manual provisioning
	if dst.Spec.IPAllocation == nil {
		t.Fatal("expected ipAllocation to be reconstructed from manual provisioning")
	}
	if dst.Spec.IPAllocation.Type != v1alpha1.IPAllocationTypeStatic {
		t.Errorf("expected type 'static', got %q", dst.Spec.IPAllocation.Type)
	}
	if dst.Spec.IPAllocation.Provider != v1alpha1.ProviderNameStaticIP {
		t.Errorf("expected provider 'static-ip-operator', got %q", dst.Spec.IPAllocation.Provider)
	}
	if dst.Spec.IPAllocation.Static == nil {
		t.Fatal("expected static config to be reconstructed")
	}
	if dst.Spec.IPAllocation.Static.IPv4CIDR != "10.0.4.0/24" {
		t.Errorf("expected ipv4CIDR '10.0.4.0/24', got %q", dst.Spec.IPAllocation.Static.IPv4CIDR)
	}
	if dst.Spec.IPAllocation.Static.IPv4Gateway != "10.0.4.1" {
		t.Errorf("expected ipv4Gateway '10.0.4.1', got %q", dst.Spec.IPAllocation.Static.IPv4Gateway)
	}
	if dst.Spec.IPAllocation.Static.VlanID != 400 {
		t.Errorf("expected vlanID 400, got %d", dst.Spec.IPAllocation.Static.VlanID)
	}
}

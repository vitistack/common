package nnconversion_test

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	crconversion "sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
	"sigs.k8s.io/randfill"

	"github.com/vitistack/common/internal/nnconversion"
	v1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	v1alpha2 "github.com/vitistack/common/pkg/v1alpha2"
	v1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
)

type spoke interface {
	conversion.Convertible
}

// TestHubRoundTrip checks hub -> spoke -> hub is lossless for random hubs.
func TestHubRoundTrip(t *testing.T) {
	spokes := map[string]func() spoke{
		"v1alpha1": func() spoke { return &v1alpha1.NetworkNamespace{} },
		"v1alpha2": func() spoke { return &v1alpha2.NetworkNamespace{} },
	}
	for name, newSpoke := range spokes {
		t.Run(name, func(t *testing.T) {
			f, r := hubFiller(1), rand.New(rand.NewPCG(1, 2))
			for i := range 2000 {
				hub := randomHub(f, r)

				sp := newSpoke()
				if err := sp.ConvertFrom(hub.DeepCopy()); err != nil {
					t.Fatalf("iteration %d: ConvertFrom: %v", i, err)
				}
				got := &v1alpha3.NetworkNamespace{}
				if err := sp.ConvertTo(got); err != nil {
					t.Fatalf("iteration %d: ConvertTo: %v", i, err)
				}
				if !equality.Semantic.DeepEqual(hub, got) {
					t.Fatalf("iteration %d: round trip changed the hub (-want +got):\n%s", i, cmp.Diff(hub, got))
				}
			}
		})
	}
}

// randomHub fills a hub, then biases open strings towards the values the
// conversions branch on. Custom fill funcs for pointer types would never
// leave the pointer nil, so this is done afterwards.
func randomHub(f *randfill.Filler, r *rand.Rand) *v1alpha3.NetworkNamespace {
	hub := &v1alpha3.NetworkNamespace{}
	f.Fill(hub)
	if p := hub.Spec.NetworkProvisioning; p != nil && r.IntN(3) > 0 {
		p.Provider = []string{nnconversion.ProviderNAM, nnconversion.ProviderManual}[r.IntN(2)]
	}
	if a := hub.Spec.IPAllocation; a != nil && r.IntN(3) > 0 {
		a.Type = []v1alpha3.IPAllocationType{v1alpha3.IPAllocationTypeDHCP, v1alpha3.IPAllocationTypeStatic}[r.IntN(2)]
	}
	return hub
}

func hubFiller(seed int64) *randfill.Filler {
	pick := func(c randfill.Continue, known ...string) string {
		if i := c.Intn(len(known) + 1); i < len(known) {
			return known[i]
		}
		return c.String(8)
	}
	return randfill.NewWithSeed(seed).NilChance(0.3).NumElements(0, 3).Funcs(
		// The conversion handler sets TypeMeta, not the conversion functions.
		func(*metav1.TypeMeta, randfill.Continue) {},
		func(m *metav1.ObjectMeta, c randfill.Continue) {
			m.Name = c.String(10)
			m.Namespace = c.String(10)
			c.Fill(&m.Labels)
			c.Fill(&m.Annotations)
			for _, k := range []string{"datacenterIdentifier", "supervisorIdentifier"} {
				if c.Bool() {
					if m.Annotations == nil {
						m.Annotations = map[string]string{}
					}
					m.Annotations[k] = pick(c, "")
				}
			}
		},
		func(s *v1alpha3.NetworkNamespaceStatus, c randfill.Continue) {
			c.FillNoCustom(s)
			if c.Bool() {
				s.ProvisioningPhase = ""
			}
			if c.Bool() {
				s.Phase = pick(c, "", "Ready", "Error")
			}
		},
		// status.egress is a list map keyed by network.
		func(es *[]v1alpha3.EgressStatus, c randfill.Continue) {
			*es = nil
			for _, n := range []string{"hnet", "inet", c.String(6)} {
				if c.Bool() {
					e := v1alpha3.EgressStatus{Network: n}
					c.Fill(&e.IPv4)
					c.Fill(&e.IPv6)
					*es = append(*es, e)
				}
			}
		},
	)
}

func TestV1alpha2ToHub(t *testing.T) {
	src := &v1alpha2.NetworkNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "ns", Annotations: map[string]string{"datacenterIdentifier": "no-west-az1"}},
		Spec: v1alpha2.NetworkNamespaceSpec{
			NetworkProvisioning: &v1alpha2.NetworkProvisioning{Provider: v1alpha2.NetworkProvisioningNAM},
			IPAllocation: &v1alpha2.NetworkNamespaceIPAllocation{
				Type: v1alpha2.IPAllocationTypeStatic,
				Static: &v1alpha2.StaticIPAllocationConfig{
					IPv4CIDR: "10.0.0.0/24", IPv4Gateway: "10.0.0.1", VlanID: 7,
					IPv4RangeStart: "10.0.0.10", DNS: []string{"1.1.1.1"},
				},
			},
		},
		Status: v1alpha2.NetworkNamespaceStatus{
			ProvisioningPhase:              v1alpha2.ProvisioningPhaseReady,
			IPv4EgressIP:                   "31.25.220.112",
			AssociatedKubernetesClusterIDs: []string{"c1"},
		},
	}

	hub := &v1alpha3.NetworkNamespace{}
	if err := src.ConvertTo(hub); err != nil {
		t.Fatal(err)
	}

	wantEgress := []v1alpha3.EgressStatus{{Network: "hnet", IPv4: []string{"31.25.220.112"}}}
	if diff := cmp.Diff(wantEgress, hub.Status.Egress); diff != "" {
		t.Errorf("status.egress (-want +got):\n%s", diff)
	}
	if hub.Status.ProvisioningPhase != v1alpha3.ProvisioningPhaseReady {
		t.Errorf("provisioningPhase = %q, want Ready", hub.Status.ProvisioningPhase)
	}
	if hub.Spec.IPAllocation.Static.IPv4RangeStart != "10.0.0.10" {
		t.Errorf("static range start not converted: %+v", hub.Spec.IPAllocation.Static)
	}
	if _, ok := hub.Annotations[nnconversion.V1alpha2DataAnnotation]; !ok {
		t.Errorf("v1alpha2-only fields (static network, cluster IDs) not stashed: %v", hub.Annotations)
	}

	back := &v1alpha2.NetworkNamespace{}
	if err := back.ConvertFrom(hub); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(src, back); diff != "" {
		t.Errorf("round trip (-want +got):\n%s", diff)
	}
}

// A v1alpha2 client changing its single egress IP replaces the hnet
// addresses but leaves other networks alone.
func TestV1alpha2EgressEdit(t *testing.T) {
	hub := &v1alpha3.NetworkNamespace{
		Spec: v1alpha3.NetworkNamespaceSpec{
			Egress: []v1alpha3.EgressRequest{{Network: "inet", DynamicIPv4: new(int32(1))}},
		},
		Status: v1alpha3.NetworkNamespaceStatus{
			Egress: []v1alpha3.EgressStatus{
				{Network: "hnet", IPv4: []string{"10.0.0.1", "10.0.0.2"}},
				{Network: "inet", IPv4: []string{"192.0.2.1"}, IPv6: []string{"2001:db8::1"}},
			},
		},
	}
	sp := &v1alpha2.NetworkNamespace{}
	if err := sp.ConvertFrom(hub); err != nil {
		t.Fatal(err)
	}
	if sp.Status.IPv4EgressIP != "10.0.0.1" || sp.Status.IPv6EgressIP != "" {
		t.Fatalf("legacy egress = %q/%q, want first hnet IPv4 only", sp.Status.IPv4EgressIP, sp.Status.IPv6EgressIP)
	}

	sp.Status.IPv4EgressIP = "10.0.0.9"
	got := &v1alpha3.NetworkNamespace{}
	if err := sp.ConvertTo(got); err != nil {
		t.Fatal(err)
	}
	want := []v1alpha3.EgressStatus{
		{Network: "hnet", IPv4: []string{"10.0.0.9"}},
		{Network: "inet", IPv4: []string{"192.0.2.1"}, IPv6: []string{"2001:db8::1"}},
	}
	if diff := cmp.Diff(want, got.Status.Egress); diff != "" {
		t.Errorf("status.egress (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(hub.Spec.Egress, got.Spec.Egress); diff != "" {
		t.Errorf("spec.egress not restored (-want +got):\n%s", diff)
	}
}

func TestV1alpha1ToHub(t *testing.T) { //nolint:gocognit,funlen
	tests := []struct {
		name string
		src  *v1alpha1.NetworkNamespace
		want func(t *testing.T, hub *v1alpha3.NetworkNamespace)
	}{
		{
			name: "no ipAllocation means NAM",
			src: &v1alpha1.NetworkNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "ns"},
				Spec:       v1alpha1.NetworkNamespaceSpec{DatacenterIdentifier: "no-west-az1", SupervisorIdentifier: "sup"},
				Status:     v1alpha1.NetworkNamespaceStatus{SupervisorIdentifier: "sup"},
			},
			want: func(t *testing.T, hub *v1alpha3.NetworkNamespace) {
				if p := hub.Spec.NetworkProvisioning; p == nil || p.Provider != "nam" {
					t.Errorf("networkProvisioning = %+v, want nam", p)
				}
				if hub.Annotations["datacenterIdentifier"] != "no-west-az1" || hub.Annotations["supervisorIdentifier"] != "sup" {
					t.Errorf("identifier annotations = %v", hub.Annotations)
				}
				if _, ok := hub.Annotations[nnconversion.V1alpha1DataAnnotation]; ok {
					t.Errorf("nothing lossy, but v1alpha1 data was stashed: %v", hub.Annotations)
				}
			},
		},
		{
			name: "static allocation means manual provisioning",
			src: &v1alpha1.NetworkNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "ns"},
				Spec: v1alpha1.NetworkNamespaceSpec{
					DatacenterIdentifier: "no-west-az1",
					IPAllocation: &v1alpha1.NetworkNamespaceIPAllocation{
						Type:     v1alpha1.IPAllocationTypeStatic,
						Provider: v1alpha1.ProviderNameStaticIP,
						Static: &v1alpha1.StaticIPAllocationConfig{
							IPv4CIDR: "10.0.2.0/24", IPv4Gateway: "10.0.2.1", VlanID: 12, IPv4RangeStart: "10.0.2.10",
						},
					},
				},
			},
			want: func(t *testing.T, hub *v1alpha3.NetworkNamespace) {
				want := &v1alpha3.NetworkProvisioning{
					Provider: "manual",
					Manual:   &v1alpha3.ManualProvisioningConfig{IPv4CIDR: "10.0.2.0/24", IPv4Gateway: "10.0.2.1", VlanID: 12},
				}
				if diff := cmp.Diff(want, hub.Spec.NetworkProvisioning); diff != "" {
					t.Errorf("networkProvisioning (-want +got):\n%s", diff)
				}
				if hub.Spec.IPAllocation.Static.IPv4RangeStart != "10.0.2.10" {
					t.Errorf("static = %+v", hub.Spec.IPAllocation.Static)
				}
			},
		},
		{
			name: "provisioningPhase is backfilled from phase",
			src: &v1alpha1.NetworkNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "ns"},
				Status:     v1alpha1.NetworkNamespaceStatus{Phase: "Ready"},
			},
			want: func(t *testing.T, hub *v1alpha3.NetworkNamespace) {
				if hub.Status.ProvisioningPhase != v1alpha3.ProvisioningPhaseReady {
					t.Errorf("provisioningPhase = %q, want Ready", hub.Status.ProvisioningPhase)
				}
			},
		},
		{
			name: "allocated IPs and cluster IDs survive via the hub",
			src: &v1alpha1.NetworkNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "ns"},
				Status: v1alpha1.NetworkNamespaceStatus{
					ProvisioningPhase:              "Ready",
					IPv4EgressIP:                   "31.25.220.112",
					AssociatedKubernetesClusterIDs: []string{"c1"},
					IPAllocationStatus: &v1alpha1.NetworkNamespaceIPAllocationStatus{
						Type: v1alpha1.IPAllocationTypeStatic, AllocatedCount: 1,
						AllocatedIPs: []v1alpha1.AllocatedIPEntry{{IP: "10.0.2.10", NetworkConfiguration: "nc"}},
					},
				},
			},
			want: func(t *testing.T, hub *v1alpha3.NetworkNamespace) {
				if ip, _ := nnconversion.LegacyEgress(hub.Status.Egress); ip != "31.25.220.112" {
					t.Errorf("hnet egress = %q", ip)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := &v1alpha3.NetworkNamespace{}
			if err := tt.src.ConvertTo(hub); err != nil {
				t.Fatal(err)
			}
			tt.want(t, hub)

			back := &v1alpha1.NetworkNamespace{}
			if err := back.ConvertFrom(hub); err != nil {
				t.Fatal(err)
			}
			want := tt.src.DeepCopy()
			if want.Status.ProvisioningPhase == "" && want.Status.Phase != "" {
				want.Status.ProvisioningPhase = string(hub.Status.ProvisioningPhase)
			}
			if want.Spec.DatacenterIdentifier != "" || want.Spec.SupervisorIdentifier != "" {
				want.Annotations = map[string]string{}
				for k, v := range map[string]string{"datacenterIdentifier": want.Spec.DatacenterIdentifier, "supervisorIdentifier": want.Spec.SupervisorIdentifier} {
					if v != "" {
						want.Annotations[k] = v
					}
				}
			}
			if diff := cmp.Diff(want, back); diff != "" {
				t.Errorf("round trip (-want +got):\n%s", diff)
			}
		})
	}
}

// TestWebhookHandler drives controller-runtime's conversion handler the way
// the API server does, including a spoke-to-spoke conversion via the hub.
func TestWebhookHandler(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
	utilruntime.Must(v1alpha3.AddToScheme(scheme))
	if ok, err := crconversion.IsConvertible(scheme, &v1alpha3.NetworkNamespace{}); err != nil || !ok {
		t.Fatalf("IsConvertible = %v, %v", ok, err)
	}
	srv := httptest.NewServer(crconversion.NewWebhookHandler(scheme, crconversion.NewRegistry()))
	defer srv.Close()

	obj := []byte(`{
		"apiVersion": "vitistack.io/v1alpha1", "kind": "NetworkNamespace",
		"metadata": {"name": "ns", "namespace": "default"},
		"spec": {"datacenterIdentifier": "no-west-az1", "supervisorIdentifier": "sup",
			"ipAllocation": {"type": "static", "static": {"ipv4CIDR": "10.0.2.0/24", "ipv4Gateway": "10.0.2.1"}}},
		"status": {"ipv4EgressIp": "31.25.220.112"}
	}`)
	review := apixv1.ConversionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "apiextensions.k8s.io/v1", Kind: "ConversionReview"},
		Request: &apixv1.ConversionRequest{
			UID:               "uid-1",
			DesiredAPIVersion: "vitistack.io/v1alpha2",
			Objects:           []runtime.RawExtension{{Raw: obj}},
		},
	}
	body, err := json.Marshal(review)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(srv.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var out apixv1.ConversionReview
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Response == nil || out.Response.Result.Status != metav1.StatusSuccess {
		t.Fatalf("conversion failed: %+v", out.Response)
	}
	if out.Response.UID != "uid-1" || len(out.Response.ConvertedObjects) != 1 {
		t.Fatalf("response = %+v", out.Response)
	}

	got := &v1alpha2.NetworkNamespace{}
	if err := json.Unmarshal(out.Response.ConvertedObjects[0].Raw, got); err != nil {
		t.Fatal(err)
	}
	if got.APIVersion != "vitistack.io/v1alpha2" || got.Kind != "NetworkNamespace" {
		t.Errorf("type meta = %v", got.TypeMeta)
	}
	if p := got.Spec.NetworkProvisioning; p == nil || p.Provider != v1alpha2.NetworkProvisioningManual || p.Manual.IPv4CIDR != "10.0.2.0/24" {
		t.Errorf("networkProvisioning = %+v", p)
	}
	if got.Spec.IPAllocation.Static.IPv4CIDR != "10.0.2.0/24" {
		t.Errorf("static network lost on v1alpha1 -> v1alpha2: %+v", got.Spec.IPAllocation.Static)
	}
	if got.Status.IPv4EgressIP != "31.25.220.112" || got.Annotations["datacenterIdentifier"] != "no-west-az1" {
		t.Errorf("status/annotations = %+v / %v", got.Status, got.Annotations)
	}
}

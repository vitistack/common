package statusapply_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/vitistack/common/pkg/operator/statusapply"
	v1alpha2 "github.com/vitistack/common/pkg/v1alpha2"
	ac "github.com/vitistack/common/pkg/v1alpha2/applyconfiguration/pkg/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var nnGVK = schema.GroupVersionKind{Group: "vitistack.io", Version: "v1alpha2", Kind: "NetworkNamespace"}

type fakeClient struct {
	client.SubResourceWriter
	calls int
	obj   runtime.ApplyConfiguration
	opts  client.SubResourceApplyOptions
}

func (c *fakeClient) GroupVersionKindFor(runtime.Object) (schema.GroupVersionKind, error) {
	return nnGVK, nil
}

func (c *fakeClient) Status() client.SubResourceWriter { return c }

func (c *fakeClient) Apply(_ context.Context, obj runtime.ApplyConfiguration, opts ...client.SubResourceApplyOption) error {
	c.calls++
	c.obj = obj
	for _, o := range opts {
		o.ApplyToSubResourceApply(&c.opts)
	}
	return nil
}

func newNN() *v1alpha2.NetworkNamespace {
	nn := &v1alpha2.NetworkNamespace{}
	nn.Name, nn.Namespace, nn.Generation = "nn1", "ns1", 7
	return nn
}

// sent decodes the request captured by the fake client.
func sent(t *testing.T, c *fakeClient) (out struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec   map[string]any                  `json:"spec"`
	Status v1alpha2.NetworkNamespaceStatus `json:"status"`
}) {
	t.Helper()
	b, err := json.Marshal(c.obj)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestApplyBatchesFieldsAndConditions(t *testing.T) {
	old := metav1.NewTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	nn := newNN()
	nn.Status.Conditions = []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue, LastTransitionTime: old}}

	p := ac.NetworkNamespaceStatusUpdater(nn)
	p.Status().WithPhase("Pending")
	p.Status().WithPhase("Ready").WithVlanID(42)
	p.SetCondition(metav1.Condition{Type: "Ready", Status: metav1.ConditionFalse, Reason: "First"})
	p.SetCondition(metav1.Condition{Type: "Synced", Status: metav1.ConditionTrue, Reason: "Ok"})
	p.SetCondition(metav1.Condition{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Success"})

	c := &fakeClient{}
	if err := p.Apply(context.Background(), c, client.FieldOwner("ctrl"), client.ForceOwnership); err != nil {
		t.Fatal(err)
	}
	if c.calls != 1 {
		t.Fatalf("calls = %d, want 1", c.calls)
	}
	if c.opts.FieldManager != "ctrl" || c.opts.Force == nil || !*c.opts.Force {
		t.Fatalf("options = %+v, want field owner ctrl and force", c.opts)
	}

	got := sent(t, c)
	if got.APIVersion != "vitistack.io/v1alpha2" || got.Kind != "NetworkNamespace" {
		t.Errorf("type = %s %s", got.APIVersion, got.Kind)
	}
	if got.Metadata.Name != "nn1" || got.Metadata.Namespace != "ns1" {
		t.Errorf("metadata = %+v", got.Metadata)
	}
	if got.Spec != nil {
		t.Errorf("spec = %v, want none", got.Spec)
	}
	if got.Status.Phase != "Ready" || got.Status.VlanID != 42 {
		t.Errorf("phase, vlan = %q, %d", got.Status.Phase, got.Status.VlanID)
	}
	if len(got.Status.Conditions) != 2 {
		t.Fatalf("conditions = %+v, want 2", got.Status.Conditions)
	}
	ready := got.Status.Conditions[0]
	if ready.Type != "Ready" || ready.Reason != "Success" || ready.ObservedGeneration != 7 {
		t.Errorf("ready = %+v", ready)
	}
	if !ready.LastTransitionTime.Equal(&old) {
		t.Errorf("ready transition time = %v, want kept %v", ready.LastTransitionTime, old)
	}
}

func TestTransitionTimeUpdatedOnStatusChange(t *testing.T) {
	old := metav1.NewTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	nn := newNN()
	nn.Status.Conditions = []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue, LastTransitionTime: old}}

	p := ac.NetworkNamespaceStatusUpdater(nn)
	p.SetCondition(metav1.Condition{Type: "Ready", Status: metav1.ConditionFalse, Reason: "Broken"})

	c := &fakeClient{}
	if err := p.Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	got := sent(t, c).Status.Conditions[0].LastTransitionTime
	if got.Time.Year() < 2026 {
		t.Errorf("transition time = %v, want now", got)
	}
}

func TestRemoveCondition(t *testing.T) {
	p := ac.NetworkNamespaceStatusUpdater(newNN())
	p.SetCondition(metav1.Condition{Type: "A", Status: metav1.ConditionTrue, Reason: "R"})
	p.SetCondition(metav1.Condition{Type: "B", Status: metav1.ConditionTrue, Reason: "R"})
	p.RemoveCondition("A")

	c := &fakeClient{}
	if err := p.Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	conds := sent(t, c).Status.Conditions
	if len(conds) != 1 || conds[0].Type != "B" {
		t.Errorf("conditions = %+v, want only B", conds)
	}
}

func TestApplySkipsWhenNoChanges(t *testing.T) {
	c := &fakeClient{}
	if err := ac.NetworkNamespaceStatusUpdater(newNN()).Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	if c.calls != 0 {
		t.Errorf("calls = %d, want 0", c.calls)
	}
}

func TestApplyRequiresFieldOwner(t *testing.T) {
	p := ac.NetworkNamespaceStatusUpdater(newNN())
	p.Status().WithPhase("Ready")
	if err := p.Apply(context.Background(), &fakeClient{}); !errors.Is(err, statusapply.ErrNoFieldOwner) {
		t.Errorf("err = %v, want statusapply.ErrNoFieldOwner", err)
	}
}

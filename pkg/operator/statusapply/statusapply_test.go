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

func updater(t *testing.T, nn *v1alpha2.NetworkNamespace) *statusapply.Patch[*ac.NetworkNamespaceStatusApplyConfiguration] {
	t.Helper()
	p, err := ac.NetworkNamespaceStatusUpdater(nn, "ctrl")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// withOwnedStatus gives nn status fields owned by two managers: "ctrl" owns
// vlanId and the Ready condition, "other" owns phase and the Synced condition.
func withOwnedStatus(nn *v1alpha2.NetworkNamespace) {
	old := metav1.NewTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	nn.Status.VlanID = 42
	nn.Status.Phase = "Ready"
	nn.Status.Conditions = []metav1.Condition{
		{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Ok", LastTransitionTime: old},
		{Type: "Synced", Status: metav1.ConditionTrue, Reason: "Ok", LastTransitionTime: old},
	}
	field := func(manager, fields string) metav1.ManagedFieldsEntry {
		return metav1.ManagedFieldsEntry{
			Manager:     manager,
			Operation:   metav1.ManagedFieldsOperationApply,
			APIVersion:  "vitistack.io/v1alpha2",
			FieldsType:  "FieldsV1",
			FieldsV1:    metav1.NewFieldsV1(fields),
			Subresource: "status",
		}
	}
	nn.ManagedFields = []metav1.ManagedFieldsEntry{
		field("ctrl", `{"f:status":{"f:vlanId":{},"f:conditions":{"k:{\"type\":\"Ready\"}":{".":{},"f:type":{},"f:status":{},"f:reason":{},"f:message":{},"f:lastTransitionTime":{}}}}}`),
		field("other", `{"f:status":{"f:phase":{},"f:conditions":{"k:{\"type\":\"Synced\"}":{".":{},"f:type":{},"f:status":{},"f:reason":{},"f:message":{},"f:lastTransitionTime":{}}}}}`),
	}
}

func TestApplyKeepsOwnedStatus(t *testing.T) {
	nn := newNN()
	withOwnedStatus(nn)

	p := updater(t, nn)
	p.SetCondition(metav1.Condition{Type: "Deleting", Status: metav1.ConditionTrue, Reason: "R"})

	c := &fakeClient{}
	if err := p.Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	got := sent(t, c).Status
	if got.VlanID != 42 {
		t.Errorf("vlanId = %d, want owned value 42 kept", got.VlanID)
	}
	if got.Phase != "" {
		t.Errorf("phase = %q, want none (owned by other)", got.Phase)
	}
	types := make([]string, 0, len(got.Conditions))
	for _, cond := range got.Conditions {
		types = append(types, cond.Type)
	}
	if len(types) != 2 || types[0] != "Ready" || types[1] != "Deleting" {
		t.Errorf("condition types = %v, want [Ready Deleting]", types)
	}
	if got.Conditions[0].Reason != "Ok" {
		t.Errorf("Ready = %+v, want kept as on the object", got.Conditions[0])
	}
}

func TestApplyOwnedStatusAloneIsNoChange(t *testing.T) {
	nn := newNN()
	withOwnedStatus(nn)

	c := &fakeClient{}
	if err := updater(t, nn).Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	if c.calls != 0 {
		t.Errorf("calls = %d, want 0", c.calls)
	}
}

func TestRemoveOwnedCondition(t *testing.T) {
	nn := newNN()
	withOwnedStatus(nn)

	p := updater(t, nn)
	p.RemoveCondition("Ready")

	c := &fakeClient{}
	if err := p.Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	got := sent(t, c).Status
	if len(got.Conditions) != 0 {
		t.Errorf("conditions = %+v, want none", got.Conditions)
	}
	if got.VlanID != 42 {
		t.Errorf("vlanId = %d, want owned value 42 kept", got.VlanID)
	}
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

	p := updater(t, nn)
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

	p := updater(t, nn)
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
	p := updater(t, newNN())
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
	if err := updater(t, newNN()).Apply(context.Background(), c, client.FieldOwner("ctrl")); err != nil {
		t.Fatal(err)
	}
	if c.calls != 0 {
		t.Errorf("calls = %d, want 0", c.calls)
	}
}

func TestApplyRequiresFieldOwner(t *testing.T) {
	p := updater(t, newNN())
	p.Status().WithPhase("Ready")
	if err := p.Apply(context.Background(), &fakeClient{}); !errors.Is(err, statusapply.ErrNoFieldOwner) {
		t.Errorf("err = %v, want statusapply.ErrNoFieldOwner", err)
	}
}

func TestGetAndHas(t *testing.T) {
	nn := newNN()
	nn.Status.Conditions = []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Old"}}
	p := updater(t, nn)

	if c, ok := p.GetCondition("Ready"); !ok || c.Reason != "Old" {
		t.Errorf("Get(Ready) = %+v, %v, want object's condition", c, ok)
	}
	if p.HasCondition("Synced") {
		t.Error("Has(Synced) = true, want false")
	}

	p.SetCondition(metav1.Condition{Type: "Ready", Status: metav1.ConditionFalse, Reason: "New"})
	p.SetCondition(metav1.Condition{Type: "Synced", Status: metav1.ConditionTrue, Reason: "Ok"})
	if c, _ := p.GetCondition("Ready"); c.Reason != "New" {
		t.Errorf("Get(Ready).Reason = %q, want recorded value New", c.Reason)
	}
	if !p.HasCondition("Synced") {
		t.Error("Has(Synced) = false, want true")
	}
}

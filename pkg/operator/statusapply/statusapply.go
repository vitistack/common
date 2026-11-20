// Package statusapply writes a resource's status subresource with Server-Side
// Apply (SSA) from any reconciler.
//
// Status changes (fields and conditions) are collected in a Patch during the
// reconcile and sent to the API server in a single apply request at the end.
//
//	patch, err := ac.NetworkNamespaceStatusUpdater(&nn, "my-controller") // generated; wraps statusapply.For
//	if err != nil {
//		return ctrl.Result{}, err
//	}
//	patch.Status().WithPhase("Ready").WithMessage("ok")
//	patch.SetCondition(&metav1.Condition{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Success"})
//	patch.SetCondition(&metav1.Condition{Type: "Synced", Status: metav1.ConditionTrue, Reason: "Success"})
//	return ctrl.Result{}, patch.Apply(ctx, r.Client,
//		client.FieldOwner("my-controller"), client.ForceOwnership)
//
// SSA semantics: each apply declares the complete set of status fields owned by
// the field manager. Fields (and conditions) this manager applied previously but
// omits now are removed unless another manager also owns them. The generated
// updaters therefore start the Patch from what the field manager already owns,
// so a reconcile only sets what changed. To remove a field, set it to nil on
// Status(); to remove a condition, use RemoveCondition.
package statusapply

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vitistack/common/pkg/operator/conditions/v2"
)

// ErrNoFieldOwner is returned by Apply when no client.FieldOwner option is given.
var ErrNoFieldOwner = errors.New("statusapply: a client.FieldOwner option is required for server-side apply")

// Client is the part of a controller-runtime client that Apply needs.
// A client.Client, or a reconciler embedding one, satisfies it.
type Client interface {
	GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error)
	Status() client.SubResourceWriter
}

// Patch accumulates status changes for a single object and writes them with
// one server-side apply call. A Patch is not safe for concurrent use.
type Patch[S any] struct {
	obj        client.Object
	status     S
	existing   []metav1.Condition
	conditions []metav1.Condition
	dirty      bool
	now        func() time.Time
}

// For starts a new Patch for obj. status is the starting status apply
// configuration: the part of obj's status the field manager owns, from the
// generated Extract<Kind>Status, or an empty one from ac.FooStatus().
// Name, namespace, generation and the current conditions are read from obj.
// The conditions in status are kept as they are on obj. Starting a Patch
// records no changes.
func For[S any](obj client.Object, status S) *Patch[S] {
	p := &Patch[S]{
		obj:      obj,
		status:   status,
		existing: currentConditions(obj),
		now:      time.Now,
	}
	for _, t := range conditionTypes(status) {
		if c, ok := conditions.Get(p.existing, t); ok {
			p.conditions = append(p.conditions, c)
		}
	}
	return p
}

// Status returns the status apply configuration so its generated With* methods
// can be chained. Calling Status marks the Patch as changed. Use SetCondition
// instead of WithConditions so conditions are de-duplicated by type.
func (p *Patch[S]) Status() S {
	p.dirty = true
	return p.status
}

// SetCondition adds or replaces (by Type) a condition in the Patch. The Patch
// stores a copy; cond itself is not modified.
// ObservedGeneration defaults to the object's generation, and LastTransitionTime
// is kept from the object's condition when its status is unchanged, otherwise
// it is set to now.
func (p *Patch[S]) SetCondition(cond *metav1.Condition) *Patch[S] {
	c := *cond
	if c.ObservedGeneration == 0 {
		c.ObservedGeneration = p.obj.GetGeneration()
	}
	if c.LastTransitionTime.IsZero() {
		c.LastTransitionTime = p.transitionTime(c.Type, c.Status)
	}
	p.dirty = true
	for i := range p.conditions {
		if p.conditions[i].Type == c.Type {
			p.conditions[i] = c
			return p
		}
	}
	p.conditions = append(p.conditions, c)
	return p
}

// RemoveCondition drops a condition of the given type from the Patch. Because of
// SSA semantics, omitting a condition this manager owned removes it from the object.
func (p *Patch[S]) RemoveCondition(condType string) *Patch[S] {
	for i := range p.conditions {
		if p.conditions[i].Type == condType {
			p.conditions = append(p.conditions[:i], p.conditions[i+1:]...)
			p.dirty = true
			return p
		}
	}
	return p
}

// GetCondition returns the condition of the given type. A condition recorded with
// SetCondition wins over the one currently on the object.
func (p *Patch[S]) GetCondition(condType string) (metav1.Condition, bool) {
	if c, ok := conditions.Get(p.conditions, condType); ok {
		return c, true
	}
	return conditions.Get(p.existing, condType)
}

// HasCondition reports whether a condition of the given type is recorded in the Patch
// or present on the object.
func (p *Patch[S]) HasCondition(condType string) bool {
	_, ok := p.GetCondition(condType)
	return ok
}

// HasChanges reports whether any status field or condition was recorded.
func (p *Patch[S]) HasChanges() bool { return p.dirty }

// Apply writes all recorded changes to the status subresource in one
// server-side apply request. c is typically r.Client. A client.FieldOwner
// option is required. Add client.ForceOwnership to take over fields still
// owned by other managers, such as legacy Update calls.
// Apply does nothing when no changes were recorded.
func (p *Patch[S]) Apply(ctx context.Context, c Client, opts ...client.SubResourceApplyOption) error {
	if !p.dirty {
		return nil
	}
	resolved := &client.SubResourceApplyOptions{}
	for _, o := range opts {
		o.ApplyToSubResourceApply(resolved)
	}
	if resolved.FieldManager == "" {
		return ErrNoFieldOwner
	}
	gvk, err := c.GroupVersionKindFor(p.obj)
	if err != nil {
		return fmt.Errorf("getting kind of %s: %w", p.obj.GetName(), err)
	}
	u, err := p.build(gvk)
	if err != nil {
		return err
	}
	if err := c.Status().Apply(ctx, client.ApplyConfigurationFromUnstructured(u), opts...); err != nil {
		return fmt.Errorf("applying status of %s %s: %w", gvk.Kind, p.obj.GetName(), err)
	}
	return nil
}

func (p *Patch[S]) build(gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	status, err := runtime.DefaultUnstructuredConverter.ToUnstructured(p.status)
	if err != nil {
		return nil, fmt.Errorf("converting status apply configuration: %w", err)
	}
	// The Patch keeps the conditions, so ones removed with RemoveCondition
	// must not come back from the starting status.
	delete(status, "conditions")
	if len(p.conditions) > 0 {
		conds := make([]any, 0, len(p.conditions))
		for i := range p.conditions {
			c, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p.conditions[i])
			if err != nil {
				return nil, fmt.Errorf("converting condition %s: %w", p.conditions[i].Type, err)
			}
			conds = append(conds, c)
		}
		status["conditions"] = conds
	}
	u := &unstructured.Unstructured{Object: map[string]any{"status": status}}
	u.SetGroupVersionKind(gvk)
	u.SetResourceVersion(p.obj.GetResourceVersion())
	u.SetName(p.obj.GetName())
	u.SetNamespace(p.obj.GetNamespace())
	return u, nil
}

func (p *Patch[S]) transitionTime(condType string, status metav1.ConditionStatus) metav1.Time {
	if e := meta.FindStatusCondition(p.existing, condType); e != nil &&
		e.Status == status && !e.LastTransitionTime.IsZero() {
		return e.LastTransitionTime
	}
	return metav1.NewTime(p.now().Truncate(time.Second))
}

// conditionTypes returns the types of the conditions in a status apply
// configuration, or nil if it has none.
func conditionTypes(status any) []string {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(status)
	if err != nil {
		return nil
	}
	var s struct {
		Conditions []struct {
			Type string `json:"type"`
		} `json:"conditions"`
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, &s); err != nil {
		return nil
	}
	types := make([]string, 0, len(s.Conditions))
	for _, c := range s.Conditions {
		types = append(types, c.Type)
	}
	return types
}

// currentConditions reads status.conditions from obj, or nil if it has none.
func currentConditions(obj client.Object) []metav1.Condition {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil
	}
	status, ok := u["status"].(map[string]any)
	if !ok {
		return nil
	}
	var s struct {
		Conditions []metav1.Condition `json:"conditions"`
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(status, &s); err != nil {
		return nil
	}
	return s.Conditions
}

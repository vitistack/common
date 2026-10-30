// Package statusapply writes a resource's status subresource with Server-Side
// Apply (SSA) from any reconciler.
//
// Status changes (fields and conditions) are collected in a Patch during the
// reconcile and sent to the API server in a single apply request at the end.
//
//	patch := ac.NetworkNamespaceStatusUpdater(&nn) // generated; wraps statusapply.For
//	patch.Status().WithPhase("Ready").WithMessage("ok")
//	patch.SetCondition(metav1.Condition{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Success"})
//	patch.SetCondition(metav1.Condition{Type: "Synced", Status: metav1.ConditionTrue, Reason: "Success"})
//	return ctrl.Result{}, patch.Apply(ctx, r.Client,
//		client.FieldOwner("my-controller"), client.ForceOwnership)
//
// SSA semantics: each apply declares the complete set of status fields owned by
// the field manager. Fields (and conditions) this manager applied previously but
// omits now are removed unless another manager also owns them. Always describe
// the full status this controller is responsible for in a single Patch.
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

// For starts a new Patch for obj. status is an empty status apply
// configuration from the generated constructor, e.g. ac.FooStatus().
// Name, namespace, generation and the current conditions are read from obj.
func For[S any](obj client.Object, status S) *Patch[S] {
	return &Patch[S]{
		obj:      obj,
		status:   status,
		existing: currentConditions(obj),
		now:      time.Now,
	}
}

// Status returns the status apply configuration so its generated With* methods
// can be chained. Calling Status marks the Patch as changed. Use SetCondition
// instead of WithConditions so conditions are de-duplicated by type.
func (p *Patch[S]) Status() S {
	p.dirty = true
	return p.status
}

// SetCondition adds or replaces (by Type) a condition in the Patch.
// ObservedGeneration defaults to the object's generation, and LastTransitionTime
// is kept from the object's condition when its status is unchanged, otherwise
// it is set to now.
func (p *Patch[S]) SetCondition(cond metav1.Condition) *Patch[S] {
	if cond.ObservedGeneration == 0 {
		cond.ObservedGeneration = p.obj.GetGeneration()
	}
	if cond.LastTransitionTime.IsZero() {
		cond.LastTransitionTime = p.transitionTime(cond)
	}
	p.dirty = true
	for i := range p.conditions {
		if p.conditions[i].Type == cond.Type {
			p.conditions[i] = cond
			return p
		}
	}
	p.conditions = append(p.conditions, cond)
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
	u.SetName(p.obj.GetName())
	u.SetNamespace(p.obj.GetNamespace())
	return u, nil
}

func (p *Patch[S]) transitionTime(cond metav1.Condition) metav1.Time {
	if e := meta.FindStatusCondition(p.existing, cond.Type); e != nil &&
		e.Status == cond.Status && !e.LastTransitionTime.IsZero() {
		return e.LastTransitionTime
	}
	return metav1.NewTime(p.now().Truncate(time.Second))
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

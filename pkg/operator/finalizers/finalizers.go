package finalizers

import (
	"context"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Has returns true if the object has the finalizer name.
func Has(obj client.Object, name string) bool {
	fins := obj.GetFinalizers()
	return slices.Contains(fins, name)
}

// Ensure adds the finalizer to the object if missing and patches it.
// Safe to call repeatedly. It performs a server-side patch to avoid conflicts.
func Ensure(ctx context.Context, c client.Client, obj client.Object, name string) error {
	if Has(obj, name) {
		return nil
	}
	// Make a shallow copy for patch base
	base := obj.DeepCopyObject().(client.Object)
	obj.SetFinalizers(append(obj.GetFinalizers(), name))
	return c.Patch(ctx, obj, client.MergeFrom(base))
}

// Remove removes the finalizer from the object if present and patches it.
// Safe to call repeatedly.
func Remove(ctx context.Context, c client.Client, obj client.Object, name string) error {
	fins := obj.GetFinalizers()
	idx := -1
	for i := range fins {
		if fins[i] == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	base := obj.DeepCopyObject().(client.Object)
	obj.SetFinalizers(append(fins[:idx], fins[idx+1:]...))
	return c.Patch(ctx, obj, client.MergeFrom(base))
}

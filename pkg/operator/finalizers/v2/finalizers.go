package finalizers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ContainsFinalizer checks if the specified finalizer is present on the object.
// Returns true if the finalizer is present, false otherwise.
func Contains(obj client.Object, finalizer string) bool {
	return controllerutil.ContainsFinalizer(obj, finalizer)
}

// Add adds the specified finalizer if missing and patches the object.
// Returns true if the finalizer was added, false otherwise.
func Add(ctx context.Context, c client.Writer, obj client.Object, finalizer string) (bool, error) {
	base := obj.DeepCopyObject().(client.Object)
	needsupdate := controllerutil.AddFinalizer(obj, finalizer)
	if needsupdate {
		err := c.Patch(ctx, obj, client.MergeFrom(base))
		return needsupdate, err
	}
	return false, nil
}

// Remove removes the specified finalizer from the object if it is present.
// Patches the object only if the finalizer is present.
func Remove(ctx context.Context, c client.Writer, obj client.Object, finalizer string) (bool, error) {
	base := obj.DeepCopyObject().(client.Object)
	needsupdate := controllerutil.RemoveFinalizer(obj, finalizer)
	if needsupdate {
		err := c.Patch(ctx, obj, client.MergeFrom(base))
		return needsupdate, err
	}
	return false, nil
}

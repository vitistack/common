package finalizers

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ContainsFinalizer checks if the specified finalizer is present on the object.
// Returns true if the finalizer is present, false otherwise.
func Contains(obj client.Object, finalizer string) bool {
	return controllerutil.ContainsFinalizer(obj, finalizer)
}

// AddFinalizer adds the specified finalizer to the object if it is not already present.
// Patches the object only if the finalizer is missing.
func Add(ctx context.Context, c client.Writer, obj client.Object, finalizer string) (ctrl.Result, error) {
	updated := controllerutil.AddFinalizer(obj, finalizer)
	if updated {
		base := obj.DeepCopyObject().(client.Object)
		if err := c.Patch(ctx, obj, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("Could not addfinalizer (patch): %w", err)
		}
	}
	return ctrl.Result{}, nil
}

// RemoveFinalizer removes the specified finalizer from the object if it is present.
// Patches the object only if the finalizer is present.
func Remove(ctx context.Context, c client.Writer, obj client.Object, finalizer string) (ctrl.Result, error) {
	updated := controllerutil.RemoveFinalizer(obj, finalizer)
	if !updated {
		return ctrl.Result{}, nil
	}
	base := obj.DeepCopyObject().(client.Object)
	return ctrl.Result{}, c.Patch(ctx, obj, client.MergeFrom(base))
}

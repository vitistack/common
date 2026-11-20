package predicates

import (
	vitistackv1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ControlPlaneVirtualSharedIPProviderPredicate admits ControlPlaneVirtualSharedIP
// events for provider. Updates pass when either the old or new object matches,
// so changing providers still triggers reconciliation.
func ControlPlaneVirtualSharedIPProviderPredicate(provider string) predicate.Predicate {
	matches := func(obj client.Object) bool {
		cp, ok := obj.(*vitistackv1alpha1.ControlPlaneVirtualSharedIP)
		return ok && cp.Spec.Provider == provider
	}
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return matches(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return matches(e.ObjectOld) || matches(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return matches(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return matches(e.Object) },
	}
}

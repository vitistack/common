package predicates

import (
	vitistackv1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NetworkNamespaceProviderPredicate admits NetworkNamespace events owned by
// the specified network provisioning provider. Updates pass when either the
// old or new object matches, so changing providers still triggers reconciliation.
func NetworkNamespaceProviderPredicate(provider string) predicate.Predicate {
	matches := func(obj client.Object) bool {
		nn, ok := obj.(*vitistackv1alpha3.NetworkNamespace)
		return ok &&
			nn.Spec.NetworkProvisioning != nil &&
			nn.Spec.NetworkProvisioning.Provider == provider
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return matches(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return matches(e.ObjectOld) || matches(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return matches(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return matches(e.Object)
		},
	}
}

package predicates

import (
	"testing"

	vitistackv1alpha2 "github.com/vitistack/common/pkg/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestNetworkNamespaceProviderPredicate(t *testing.T) {
	t.Parallel()

	networkNamespace := func(provider vitistackv1alpha2.NetworkProvisioningType) *vitistackv1alpha2.NetworkNamespace {
		return &vitistackv1alpha2.NetworkNamespace{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			Spec: vitistackv1alpha2.NetworkNamespaceSpec{
				NetworkProvisioning: &vitistackv1alpha2.NetworkProvisioning{Provider: provider},
			},
		}
	}

	const selectedProvider = vitistackv1alpha2.NetworkProvisioningNAM
	otherProvider := vitistackv1alpha2.NetworkProvisioningType("other")
	filter := NetworkNamespaceProviderPredicate(selectedProvider)

	tests := []struct {
		name string
		old  *vitistackv1alpha2.NetworkNamespace
		new  *vitistackv1alpha2.NetworkNamespace
		want bool
	}{
		{name: "update remains selected", old: networkNamespace(selectedProvider), new: networkNamespace(selectedProvider), want: true},
		{name: "update becomes selected", old: networkNamespace(otherProvider), new: networkNamespace(selectedProvider), want: true},
		{name: "update stops being selected", old: networkNamespace(selectedProvider), new: networkNamespace(otherProvider), want: true},
		{name: "update unrelated provider", old: networkNamespace(otherProvider), new: networkNamespace(otherProvider), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := filter.Update(event.UpdateEvent{ObjectOld: tt.old, ObjectNew: tt.new})
			if got != tt.want {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("create and delete match selected provider", func(t *testing.T) {
		t.Parallel()
		nn := networkNamespace(selectedProvider)
		if !filter.Create(event.CreateEvent{Object: nn}) {
			t.Error("Create() = false, want true")
		}
		if !filter.Delete(event.DeleteEvent{Object: nn}) {
			t.Error("Delete() = false, want true")
		}
	})

	t.Run("unset provisioning does not match", func(t *testing.T) {
		t.Parallel()
		nn := &vitistackv1alpha2.NetworkNamespace{}
		if filter.Create(event.CreateEvent{Object: nn}) {
			t.Error("Create() = true, want false")
		}
	})
}

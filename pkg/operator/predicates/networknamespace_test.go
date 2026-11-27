package predicates

import (
	"testing"

	vitistackv1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestNetworkNamespaceProviderPredicate(t *testing.T) {
	t.Parallel()

	networkNamespace := func(provider string) *vitistackv1alpha3.NetworkNamespace {
		return &vitistackv1alpha3.NetworkNamespace{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			Spec: vitistackv1alpha3.NetworkNamespaceSpec{
				NetworkProvisioning: &vitistackv1alpha3.NetworkProvisioning{Provider: provider},
			},
		}
	}

	const selectedProvider = "nam"
	otherProvider := "other"
	filter := NetworkNamespaceProviderPredicate(selectedProvider)

	tests := []struct {
		name string
		old  *vitistackv1alpha3.NetworkNamespace
		new  *vitistackv1alpha3.NetworkNamespace
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
		nn := &vitistackv1alpha3.NetworkNamespace{}
		if filter.Create(event.CreateEvent{Object: nn}) {
			t.Error("Create() = true, want false")
		}
	})
}

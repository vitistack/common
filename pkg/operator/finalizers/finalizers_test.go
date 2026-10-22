package finalizers

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	finalizer1Name    = "finalizer1"
	testFinalizerName = "test-finalizer"
	mockObjectKind    = "MockObject"
	testObjectName    = "test-object"
	defaultNamespace  = "default"
	firstFinalizer    = "first"
)

// mockObject is a simple test object that implements client.Object
type mockObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (m *mockObject) DeepCopyObject() runtime.Object {
	out := &mockObject{
		TypeMeta: m.TypeMeta,
	}
	m.DeepCopyInto(&out.ObjectMeta)
	return out
}

// mockObjectList is needed for scheme registration
type mockObjectList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []mockObject
}

func (m *mockObjectList) DeepCopyObject() runtime.Object {
	out := &mockObjectList{
		TypeMeta: m.TypeMeta,
	}
	m.DeepCopyInto(&out.ListMeta)
	if m.Items != nil {
		out.Items = make([]mockObject, len(m.Items))
		for i := range m.Items {
			out.Items[i] = *m.Items[i].DeepCopyObject().(*mockObject)
		}
	}
	return out
}

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: "", Version: "v1"}
	scheme.AddKnownTypes(gv, &mockObject{}, &mockObjectList{})
	return scheme
}

func TestHas(t *testing.T) {
	tests := []struct {
		name          string
		finalizers    []string
		finalizerName string
		expected      bool
	}{
		{
			name:          "has finalizer",
			finalizers:    []string{finalizer1Name, "finalizer2"},
			finalizerName: finalizer1Name,
			expected:      true,
		},
		{
			name:          "does not have finalizer",
			finalizers:    []string{finalizer1Name, "finalizer2"},
			finalizerName: "finalizer3",
			expected:      false,
		},
		{
			name:          "empty finalizers list",
			finalizers:    []string{},
			finalizerName: finalizer1Name,
			expected:      false,
		},
		{
			name:          "nil finalizers list",
			finalizers:    nil,
			finalizerName: finalizer1Name,
			expected:      false,
		},
		{
			name:          "has finalizer at end",
			finalizers:    []string{"a", "b", "c"},
			finalizerName: "c",
			expected:      true,
		},
		{
			name:          "has finalizer in middle",
			finalizers:    []string{"a", "b", "c"},
			finalizerName: "b",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &mockObject{}
			obj.SetFinalizers(tt.finalizers)

			result := Has(obj, tt.finalizerName)
			if result != tt.expected {
				t.Errorf("Has() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnsure(t *testing.T) {
	tests := []struct {
		name               string
		initialFinalizers  []string
		finalizerName      string
		expectError        bool
		expectedFinalizers []string
	}{
		{
			name:               "adds finalizer when not present",
			initialFinalizers:  []string{},
			finalizerName:      testFinalizerName,
			expectError:        false,
			expectedFinalizers: []string{testFinalizerName},
		},
		{
			name:               "does not duplicate when already present",
			initialFinalizers:  []string{testFinalizerName},
			finalizerName:      testFinalizerName,
			expectError:        false,
			expectedFinalizers: []string{testFinalizerName},
		},
		{
			name:               "adds to existing finalizers",
			initialFinalizers:  []string{"existing-finalizer"},
			finalizerName:      "new-finalizer",
			expectError:        false,
			expectedFinalizers: []string{"existing-finalizer", "new-finalizer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &mockObject{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       mockObjectKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       testObjectName,
					Namespace:  defaultNamespace,
					Finalizers: tt.initialFinalizers,
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newTestScheme()).
				WithObjects(obj).
				Build()

			ctx := context.Background()
			err := Ensure(ctx, fakeClient, obj, tt.finalizerName)

			if tt.expectError && err == nil {
				t.Errorf("Ensure() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Ensure() unexpected error: %v", err)
			}

			// When finalizer already exists, the finalizers should remain unchanged
			if !tt.expectError {
				finalizers := obj.GetFinalizers()
				if len(finalizers) != len(tt.expectedFinalizers) {
					t.Errorf("Ensure() finalizers length = %d, want %d", len(finalizers), len(tt.expectedFinalizers))
				}

				// Check that the finalizer is present
				if !Has(obj, tt.finalizerName) {
					t.Errorf("Ensure() finalizer %q should be present", tt.finalizerName)
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name               string
		initialFinalizers  []string
		finalizerName      string
		expectError        bool
		expectedFinalizers []string
	}{
		{
			name:               "removes existing finalizer",
			initialFinalizers:  []string{testFinalizerName},
			finalizerName:      testFinalizerName,
			expectError:        false,
			expectedFinalizers: []string{},
		},
		{
			name:               "no-op when finalizer not present",
			initialFinalizers:  []string{"other-finalizer"},
			finalizerName:      testFinalizerName,
			expectError:        false,
			expectedFinalizers: []string{"other-finalizer"},
		},
		{
			name:               "removes from multiple finalizers",
			initialFinalizers:  []string{"a", "b", "c"},
			finalizerName:      "b",
			expectError:        false,
			expectedFinalizers: []string{"a", "c"},
		},
		{
			name:               "removes first finalizer",
			initialFinalizers:  []string{firstFinalizer, "second"},
			finalizerName:      firstFinalizer,
			expectError:        false,
			expectedFinalizers: []string{"second"},
		},
		{
			name:               "removes last finalizer",
			initialFinalizers:  []string{firstFinalizer, "last"},
			finalizerName:      "last",
			expectError:        false,
			expectedFinalizers: []string{firstFinalizer},
		},
		{
			name:               "no-op on empty list",
			initialFinalizers:  []string{},
			finalizerName:      testFinalizerName,
			expectError:        false,
			expectedFinalizers: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &mockObject{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       mockObjectKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       testObjectName,
					Namespace:  defaultNamespace,
					Finalizers: tt.initialFinalizers,
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newTestScheme()).
				WithObjects(obj).
				Build()

			ctx := context.Background()
			err := Remove(ctx, fakeClient, obj, tt.finalizerName)

			if tt.expectError && err == nil {
				t.Errorf("Remove() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Remove() unexpected error: %v", err)
			}

			if !tt.expectError {
				finalizers := obj.GetFinalizers()
				if len(finalizers) != len(tt.expectedFinalizers) {
					t.Errorf("Remove() finalizers = %v, want %v", finalizers, tt.expectedFinalizers)
				}

				// Verify the finalizer is actually removed
				if Has(obj, tt.finalizerName) {
					t.Errorf("Remove() finalizer %q should not be present", tt.finalizerName)
				}
			}
		})
	}
}

func TestEnsureIdempotent(t *testing.T) {
	obj := &mockObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       mockObjectKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testObjectName,
			Namespace: defaultNamespace,
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(newTestScheme()).
		WithObjects(obj).
		Build()

	ctx := context.Background()
	finalizerName := testFinalizerName

	// Call Ensure multiple times
	for i := range 3 {
		err := Ensure(ctx, fakeClient, obj, finalizerName)
		if err != nil {
			t.Fatalf("Ensure() iteration %d unexpected error: %v", i, err)
		}

		if !Has(obj, finalizerName) {
			t.Errorf("Ensure() iteration %d: finalizer should be present", i)
		}
	}

	// Should only have one instance of the finalizer
	count := 0
	for _, f := range obj.GetFinalizers() {
		if f == finalizerName {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Ensure() should add finalizer only once, found %d instances", count)
	}
}

func TestRemoveIdempotent(t *testing.T) {
	obj := &mockObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       mockObjectKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       testObjectName,
			Namespace:  "default",
			Finalizers: []string{testFinalizerName},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(newTestScheme()).
		WithObjects(obj).
		Build()

	ctx := context.Background()
	finalizerName := testFinalizerName

	// Call Remove multiple times
	for i := range 3 {
		err := Remove(ctx, fakeClient, obj, finalizerName)
		if err != nil {
			t.Fatalf("Remove() iteration %d unexpected error: %v", i, err)
		}

		if Has(obj, finalizerName) {
			t.Errorf("Remove() iteration %d: finalizer should not be present", i)
		}
	}

	// Finalizer should be completely gone
	if len(obj.GetFinalizers()) != 0 {
		t.Errorf("Remove() should remove all instances, got finalizers: %v", obj.GetFinalizers())
	}
}

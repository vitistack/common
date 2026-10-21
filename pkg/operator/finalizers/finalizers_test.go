package finalizers

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type testTraceCollector struct {
	collectortrace.UnimplementedTraceServiceServer
}

func (testTraceCollector) Export(context.Context, *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

type finalizerWriter struct {
	client.Writer
	patches int
	err     error
}

func (w *finalizerWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	w.patches++
	if w.err != nil {
		return w.err
	}
	return w.Writer.Patch(ctx, obj, patch, opts...)
}

func (w *finalizerWriter) Update(context.Context, client.Object, ...client.UpdateOption) error {
	return errors.New("expected Patch, not Update")
}

func TestFinalizerPatches(t *testing.T) {
	patchErr := errors.New("patch failed")
	for _, tt := range []struct {
		name    string
		run     func(context.Context, client.Writer, client.Object, string) (ctrl.Result, error)
		initial []string
		want    []string
		patches int
		err     error
	}{
		{name: "add", run: Add, initial: []string{"other"}, want: []string{"other", "target"}, patches: 1},
		{name: "already present", run: Add, initial: []string{"target"}, want: []string{"target"}},
		{name: "remove", run: Remove, initial: []string{"target", "other", "target"}, want: []string{"other"}, patches: 1},
		{name: "already absent", run: Remove, initial: []string{"other"}, want: []string{"other"}},
		{name: "add error", run: Add, patches: 1, err: patchErr},
		{name: "remove error", run: Remove, initial: []string{"target"}, want: []string{"target"}, patches: 1, err: patchErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stored := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: "default", Finalizers: tt.initial,
				Labels: map[string]string{"version": "old"},
			}}
			c := fake.NewClientBuilder().WithObjects(stored).Build()
			obj := &corev1.ConfigMap{}
			key := client.ObjectKeyFromObject(stored)
			if err := c.Get(ctx, key, obj); err != nil {
				t.Fatal(err)
			}

			// Simulate an unrelated change after the reconciler read the object.
			latest := obj.DeepCopy()
			latest.Labels["version"] = "new"
			if err := c.Update(ctx, latest); err != nil {
				t.Fatal(err)
			}

			w := &finalizerWriter{Writer: c, err: tt.err}
			result, err := tt.run(ctx, w, obj, "target")
			if !errors.Is(err, tt.err) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
			if result != (ctrl.Result{}) {
				t.Fatalf("unexpected result: %v", result)
			}
			if w.patches != tt.patches {
				t.Fatalf("patch calls = %d, want %d", w.patches, tt.patches)
			}
			if err := c.Get(ctx, key, stored); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(stored.Finalizers, tt.want) {
				t.Fatalf("finalizers = %v, want %v", stored.Finalizers, tt.want)
			}
			if stored.Labels["version"] != "new" {
				t.Fatal("patch overwrote an unrelated label change")
			}
		})
	}
}

func TestContains(t *testing.T) {
	for _, tt := range []struct {
		name       string
		finalizers []string
		want       bool
	}{
		{name: "nil finalizers"},
		{name: "empty finalizers", finalizers: []string{}},
		{name: "absent", finalizers: []string{"other"}},
		{name: "present", finalizers: []string{"target"}, want: true},
		{name: "present among others", finalizers: []string{"first", "target", "last"}, want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			obj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Finalizers: tt.finalizers}}
			before := obj.DeepCopy()
			if got := Contains(obj, "target"); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(obj, before) {
				t.Error("Contains modified the object")
			}
		})
	}
}

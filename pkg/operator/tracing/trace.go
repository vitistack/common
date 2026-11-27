package tracing

import (
	"context"

	"github.com/vitistack/common/pkg/tracing/vititracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// Start sets up logging and tracing for a reconciliation request.
// Returns a context containing the span, and the span itself.
func Start(ctx context.Context, name string, req ctrl.Request) (context.Context, trace.Span) {
	ctx, span := vititracer.StartSpan(ctx, name,
		trace.WithAttributes(
			attribute.String("k8s.namespace", req.Namespace),
			attribute.String("k8s.name", req.Name),
			attribute.String("reconcile.id", string(controller.ReconcileIDFromContext(ctx))),
		))
	return ctx, span
}

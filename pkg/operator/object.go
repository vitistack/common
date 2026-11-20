package operator

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetObject(ctx context.Context, r client.Reader, req ctrl.Request, obj client.Object) error {
	logger := log.FromContext(ctx)
	span := trace.SpanFromContext(ctx)
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			span.SetStatus(codes.Ok, "Resource not found")
			logger.V(1).Info("Resource not found", "namespace", req.Namespace, "name", req.Name)
			return nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error(err, "Failed to get object", "namespace", req.Namespace, "name", req.Name)
		return err
	}
	span.SetAttributes(
		attribute.Int64("k8s.generation", obj.GetGeneration()),
	)
	return nil
}

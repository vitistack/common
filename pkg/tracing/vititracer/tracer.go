package vititracer

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

var (
	globalTracer *tracer
	initOnce     sync.Once
	initErr      error
)

type tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// Init initializes the global tracer singleton. It is safe to call multiple
// times; only the first call takes effect. Subsequent calls return the error
// (if any) from the first initialization.
func Init(ctx context.Context, serviceName string, opts ...Option) error {
	initOnce.Do(func() {
		initErr = initTracer(ctx, serviceName, opts...)
	})
	return initErr
}

func initTracer(ctx context.Context, serviceName string, opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	resourceOpts := []resource.Option{resource.WithFromEnv()}
	if serviceName != "" {
		resourceOpts = append(resourceOpts, resource.WithAttributes(semconv.ServiceName(serviceName)))
	}
	res, err := resource.New(ctx, resourceOpts...)
	if err != nil {
		return fmt.Errorf("trace: failed to create resource: %w", err)
	}

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithTimeout(cfg.timeout),
	}
	if cfg.endpoint != "" {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithEndpoint(cfg.endpoint))
	}
	if cfg.retry != nil {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithRetry(*cfg.retry))
	}
	if cfg.insecure != nil || cfg.endpoint != "" {
		if cfg.isInsecure() {
			exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
		} else {
			exporterOpts = append(exporterOpts, otlptracegrpc.WithTLSCredentials(
				credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}),
			))
		}
	}

	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return fmt.Errorf("trace: failed to create exporter: %w", err)
	}

	providerOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	}
	if cfg.sampler != nil {
		providerOpts = append(providerOpts, sdktrace.WithSampler(cfg.sampler))
	}
	provider := sdktrace.NewTracerProvider(providerOpts...)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	globalTracer = &tracer{
		provider: provider,
		tracer:   provider.Tracer(serviceName),
	}
	return nil
}

// Shutdown gracefully shuts down the tracer provider, flushing any pending spans.
func Shutdown(ctx context.Context) error {
	if globalTracer == nil {
		return nil
	}
	return globalTracer.provider.Shutdown(ctx)
}

// InitWithDefault initializes tracing unless OTEL_SDK_DISABLED is true.
// Service name, resource attributes, OTLP endpoint/security, and sampling are
// configured through standard OTEL environment variables. Explicit options
// override their corresponding environment settings. TRACING_ENABLED is ignored.
func InitWithDefault(ctx context.Context, opts ...Option) error {
	if value := os.Getenv("OTEL_SDK_DISABLED"); value != "" {
		disabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("trace: invalid OTEL_SDK_DISABLED: %w", err)
		}
		if disabled {
			return nil
		}
	}
	return Init(ctx, "", opts...)
}

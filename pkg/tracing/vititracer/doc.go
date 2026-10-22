// Package vititracer provides process-wide OpenTelemetry tracing with a batched
// OTLP/gRPC exporter and W3C Trace Context and Baggage propagation.
//
// # Startup and shutdown
//
// Call InitWithDefault once during application startup, before starting workers.
// Environment variables configure tracing; setting them alone does not initialize
// the package. Handle initialization errors and call Shutdown with a fresh,
// bounded context before the application exits to flush buffered spans.
//
//	if err := vititracer.InitWithDefault(ctx); err != nil {
//	    return err
//	}
//	defer func() {
//	    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	    defer cancel()
//	    if err := vititracer.Shutdown(shutdownCtx); err != nil {
//	        log.Printf("shutting down tracing: %v", err)
//	    }
//	}()
//
// Initialization is process-wide and takes effect only once. Changing the
// environment afterward does not reconfigure the tracer, and Shutdown does not
// allow it to be initialized again. Successful initialization does not guarantee
// that the collector is reachable; export happens asynchronously.
//
// # Recommended environment configuration
//
// Configure a service name and a reachable collector with an OTLP/gRPC receiver:
//
//	OTEL_SERVICE_NAME=nms-operator
//	OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
//	OTEL_TRACES_SAMPLER=parentbased_always_on
//	OTEL_RESOURCE_ATTRIBUTES=deployment.environment.name=production
//
// The endpoint is a URL. Use http for an unencrypted connection or https for TLS.
// This package always uses OTLP/gRPC, usually on port 4317; setting
// OTEL_EXPORTER_OTLP_PROTOCOL does not switch it to OTLP/HTTP on port 4318.
//
// Supported environment settings include:
//
//   - OTEL_SDK_DISABLED: true makes InitWithDefault skip initialization. Unset or
//     false enables initialization. Invalid boolean values return an error.
//     TRACING_ENABLED is no longer consulted. This switch does not shut down an
//     already initialized tracer, and calling Init directly bypasses it.
//   - OTEL_SERVICE_NAME: identifies the application in traces. It takes precedence
//     over service.name in OTEL_RESOURCE_ATTRIBUTES. Set it explicitly; this
//     package does not derive a service name from ROLE or NAME.
//   - OTEL_RESOURCE_ATTRIBUTES: comma-separated resource attributes, such as
//     service.version=1.2.3,deployment.environment.name=production.
//   - OTEL_EXPORTER_OTLP_ENDPOINT: collector URL; the exporter defaults to
//     https://localhost:4317 when no endpoint is configured.
//   - OTEL_EXPORTER_OTLP_TRACES_ENDPOINT: trace-specific endpoint, taking
//     precedence over OTEL_EXPORTER_OTLP_ENDPOINT.
//   - OTEL_EXPORTER_OTLP_HEADERS and OTEL_EXPORTER_OTLP_TRACES_HEADERS:
//     authentication or other exporter headers. Trace-specific configuration
//     takes precedence. Supply credentials through a secret, not source code.
//   - OTEL_EXPORTER_OTLP_CERTIFICATE and OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE:
//     paths to a trusted PEM certificate for TLS.
//   - OTEL_TRACES_SAMPLER: sampling policy. The default is
//     parentbased_always_on. Use always_off to record no spans, or
//     parentbased_traceidratio with OTEL_TRACES_SAMPLER_ARG between 0 and 1 to
//     sample a fraction of root traces while respecting parent decisions.
//
// # Explicit overrides and limitations
//
// Options passed to InitWithDefault override their corresponding environment
// settings. Init additionally accepts an explicit service name, which overrides
// the environment when nonempty.
//
// WithEndpoint takes host:port, not a URL. For this explicit endpoint option,
// the package enables TLS on port 443 and otherwise defaults to an unencrypted
// connection. WithInsecure explicitly disables TLS. Prefer an environment
// endpoint URL when selecting TLS by scheme.
//
// The package sets a two-second export timeout, overriding
// OTEL_EXPORTER_OTLP_TIMEOUT and OTEL_EXPORTER_OTLP_TRACES_TIMEOUT; use WithTimeout
// to change it. WithSampler overrides OTEL_TRACES_SAMPLER. The exporter and
// propagators are fixed: OTEL_TRACES_EXPORTER and OTEL_PROPAGATORS do not select
// alternatives. This package configures traces only, not metrics or logs.
//
// # Creating spans
//
// StartSpan returns a context containing the new span. Pass that context to
// downstream work and always end the span:
//
//	ctx, span := vititracer.StartSpan(ctx, "reconcile")
//	defer span.End()
//
// Use SpanError to record an error and mark the span as failed, and SpanOk to
// mark success. StartSpan is safe before initialization or when initialization
// is disabled: it returns a non-recording span. SuppressTracing disables span
// recording for a context and its descendants, for example for health checks.
package vititracer

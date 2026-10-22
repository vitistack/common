package vititracer

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

type testCollector struct {
	collectortrace.UnimplementedTraceServiceServer
	requests chan *collectortrace.ExportTraceServiceRequest
}

func (c *testCollector) Export(_ context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	c.requests <- req
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func startTestCollector(t *testing.T) (string, *testCollector) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	collector := &testCollector{requests: make(chan *collectortrace.ExportTraceServiceRequest, 1)}
	server := grpc.NewServer()
	collectortrace.RegisterTraceServiceServer(server, collector)
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		if err := <-serveErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("serving collector: %v", err)
		}
	})
	return "http://" + listener.Addr().String(), collector
}

func cleanupTestTracer(t *testing.T) {
	t.Helper()
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		if globalTracer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := Shutdown(ctx); err != nil {
				t.Errorf("shutting down tracer: %v", err)
			}
		}
		globalTracer = nil
		initOnce = sync.Once{}
		initErr = nil
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
}

func assertExportedResource(t *testing.T, collector *testCollector) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := globalTracer.provider.ForceFlush(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case req := <-collector.requests:
		if len(req.ResourceSpans) != 1 {
			t.Fatalf("resource span groups = %d, want 1", len(req.ResourceSpans))
		}
		attrs := map[string]string{}
		for _, attr := range req.ResourceSpans[0].Resource.Attributes {
			attrs[attr.Key] = attr.Value.GetStringValue()
		}
		if attrs["service.name"] != "test-service" || attrs["deployment.environment.name"] != "test" {
			t.Errorf("unexpected resource attributes: %v", attrs)
		}
	case <-ctx.Done():
		t.Fatal("no spans exported to the environment-configured endpoint")
	}
}

func TestInitWithDefault(t *testing.T) {
	for _, tt := range []struct {
		name        string
		disabled    string
		sampler     string
		override    bool
		wantInit    bool
		wantSampled bool
		wantError   bool
	}{
		{name: "enabled by default", sampler: "always_on", wantInit: true, wantSampled: true},
		{name: "explicitly enabled", disabled: "false", sampler: "always_on", wantInit: true, wantSampled: true},
		{name: "disabled", disabled: "true"},
		{name: "invalid disable value", disabled: "invalid", wantError: true},
		{name: "environment sampler", sampler: "always_off", wantInit: true},
		{name: "explicit sampler override", sampler: "always_off", override: true, wantInit: true, wantSampled: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, collector := startTestCollector(t)
			cleanupTestTracer(t)
			t.Setenv("OTEL_SDK_DISABLED", tt.disabled)
			t.Setenv("TRACING_ENABLED", "false")
			t.Setenv("OTEL_SERVICE_NAME", "test-service")
			t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment.name=test")
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", endpoint)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
			t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "")
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE", "")
			t.Setenv("OTEL_TRACES_SAMPLER", tt.sampler)
			t.Setenv("OTEL_TRACES_SAMPLER_ARG", "")
			var opts []Option
			if tt.override {
				opts = append(opts, WithSampler(sdktrace.AlwaysSample()))
			}
			err := InitWithDefault(context.Background(), opts...)
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "OTEL_SDK_DISABLED") {
					t.Fatalf("expected a configuration error, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if (globalTracer != nil) != tt.wantInit {
				t.Fatalf("tracer initialized = %v, want %v", globalTracer != nil, tt.wantInit)
			}
			_, span := StartSpan(context.Background(), "test")
			if span.IsRecording() != tt.wantSampled {
				t.Errorf("recording = %v, want %v", span.IsRecording(), tt.wantSampled)
			}
			span.End()
			if tt.wantSampled {
				assertExportedResource(t, collector)
			}
		})
	}
}

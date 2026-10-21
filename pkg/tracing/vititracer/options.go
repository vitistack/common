package vititracer

import (
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type traceConfig struct {
	endpoint string
	insecure *bool // nil = auto-detect from endpoint
	sampler  sdktrace.Sampler
	timeout  time.Duration
	retry    *otlptracegrpc.RetryConfig
}

func defaultConfig() traceConfig {
	return traceConfig{
		timeout: 2 * time.Second,
		retry: &otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			MaxElapsedTime:  20 * time.Second,
		},
	}
}

func (c traceConfig) isInsecure() bool {
	if c.insecure != nil {
		return *c.insecure
	}
	return !strings.HasSuffix(c.endpoint, ":443")
}

// Option configures the tracer initialization.
type Option func(*traceConfig)

// WithEndpoint sets the OTLP gRPC collector endpoint (e.g. "localhost:4317").
func WithEndpoint(endpoint string) Option {
	return func(c *traceConfig) {
		c.endpoint = endpoint
	}
}

// WithInsecure disables TLS for the gRPC connection to the collector.
// With an explicit WithEndpoint, TLS is disabled unless the endpoint port is 443.
// Otherwise, endpoint security is configured by the OTLP environment settings.
func WithInsecure() Option {
	return func(c *traceConfig) {
		insecure := true
		c.insecure = &insecure
	}
}

// WithSampler overrides OTEL_TRACES_SAMPLER and its default parentbased_always_on.
func WithSampler(sampler sdktrace.Sampler) Option {
	return func(c *traceConfig) {
		c.sampler = sampler
	}
}

// WithTimeout overrides the default export timeout (2s).
func WithTimeout(d time.Duration) Option {
	return func(c *traceConfig) {
		c.timeout = d
	}
}

// WithRetry overrides the default retry configuration.
// Pass nil to disable retries.
func WithRetry(cfg *otlptracegrpc.RetryConfig) Option {
	return func(c *traceConfig) {
		c.retry = cfg
	}
}

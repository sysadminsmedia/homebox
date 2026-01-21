// Package otel provides OpenTelemetry tracing initialization and utilities.
package otel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider wraps the OpenTelemetry TracerProvider with additional functionality.
type Provider struct {
	tp       *sdktrace.TracerProvider
	cfg      *config.OTelConfig
	shutdown func(context.Context) error
}

// TracerProvider returns the underlying OpenTelemetry TracerProvider.
func (p *Provider) TracerProvider() trace.TracerProvider {
	if p == nil || p.tp == nil {
		return otel.GetTracerProvider()
	}
	return p.tp
}

// Tracer returns a named tracer from this provider.
func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return p.TracerProvider().Tracer(name, opts...)
}

// Shutdown gracefully shuts down the trace provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.shutdown == nil {
		return nil
	}
	return p.shutdown(ctx)
}

// IsEnabled returns true if OpenTelemetry tracing is enabled.
func (p *Provider) IsEnabled() bool {
	return p != nil && p.cfg != nil && p.cfg.Enabled
}

// Config returns the OpenTelemetry configuration.
func (p *Provider) Config() *config.OTelConfig {
	if p == nil {
		return nil
	}
	return p.cfg
}

// NewProvider creates a new OpenTelemetry provider based on the configuration.
// If OTel is disabled, returns a no-op provider that can still be used safely.
func NewProvider(ctx context.Context, cfg *config.OTelConfig, buildVersion string) (*Provider, error) {
	if cfg == nil || !cfg.Enabled {
		log.Debug().Msg("OpenTelemetry tracing is disabled")
		return &Provider{cfg: cfg}, nil
	}

	// Create resource with service information
	serviceVersion := cfg.ServiceVersion
	if serviceVersion == "" {
		serviceVersion = buildVersion
	}

	// Note: we don't merge with resource.Default() to avoid schema URL conflicts
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(serviceVersion),
		attribute.String("deployment.environment", getEnvironment()),
	)

	// Create exporter based on configuration
	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Configure sampler
	var sampler sdktrace.Sampler
	switch {
	case cfg.SampleRate >= 1.0:
		sampler = sdktrace.AlwaysSample()
	case cfg.SampleRate <= 0:
		sampler = sdktrace.NeverSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sampler)),
	)

	// Set as global provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Info().
		Str("service", cfg.ServiceName).
		Str("version", serviceVersion).
		Str("exporter", cfg.Exporter).
		Str("endpoint", cfg.Endpoint).
		Float64("sample_rate", cfg.SampleRate).
		Msg("OpenTelemetry tracing initialized")

	return &Provider{
		tp:  tp,
		cfg: cfg,
		shutdown: func(ctx context.Context) error {
			// Give pending spans time to be exported
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return tp.Shutdown(ctx)
		},
	}, nil
}

// createExporter creates the appropriate exporter based on configuration.
func createExporter(ctx context.Context, cfg *config.OTelConfig) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(cfg.Exporter) {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())

	case "none", "":
		return &noopExporter{}, nil

	case "otlp":
		return createOTLPExporter(ctx, cfg)

	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", cfg.Exporter)
	}
}

// createOTLPExporter creates an OTLP exporter based on the protocol configuration.
func createOTLPExporter(ctx context.Context, cfg *config.OTelConfig) (*otlptrace.Exporter, error) {
	headers := parseHeaders(cfg.Headers)

	switch strings.ToLower(cfg.Protocol) {
	case "http":
		opts := []otlptracehttp.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if len(headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(headers))
		}
		return otlptracehttp.New(ctx, opts...)

	case "grpc", "":
		opts := []otlptracegrpc.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(cfg.Endpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
			opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		}
		if len(headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(headers))
		}
		return otlptracegrpc.New(ctx, opts...)

	default:
		return nil, fmt.Errorf("unsupported OTLP protocol: %s", cfg.Protocol)
	}
}

// parseHeaders parses a comma-separated list of key=value pairs into a map.
func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

// getEnvironment returns the current deployment environment.
// ...existing code...
func getEnvironment() string {
	return "production"
}

// noopExporter is a no-op span exporter for when tracing is disabled.
type noopExporter struct{}

func (e *noopExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *noopExporter) Shutdown(_ context.Context) error {
	return nil
}

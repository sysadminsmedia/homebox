// Package otel provides OpenTelemetry tracing, metrics, and logging initialization and utilities.
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
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider wraps the OpenTelemetry providers with additional functionality.
type Provider struct {
	tp       *sdktrace.TracerProvider
	mp       *sdkmetric.MeterProvider
	lp       *sdklog.LoggerProvider
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

// MeterProvider returns the underlying OpenTelemetry MeterProvider.
func (p *Provider) MeterProvider() metric.MeterProvider {
	if p == nil || p.mp == nil {
		return otel.GetMeterProvider()
	}
	return p.mp
}

// LoggerProvider returns the underlying OpenTelemetry LoggerProvider.
func (p *Provider) LoggerProvider() otellog.LoggerProvider {
	if p == nil || p.lp == nil {
		return global.GetLoggerProvider()
	}
	return p.lp
}

// Tracer returns a named tracer from this provider.
func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return p.TracerProvider().Tracer(name, opts...)
}

// Meter returns a named meter from this provider.
func (p *Provider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return p.MeterProvider().Meter(name, opts...)
}

// Logger returns a named logger from this provider.
func (p *Provider) Logger(name string, opts ...otellog.LoggerOption) otellog.Logger {
	return p.LoggerProvider().Logger(name, opts...)
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
		log.Debug().Msg("OpenTelemetry is disabled")
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

	var tp *sdktrace.TracerProvider
	var mp *sdkmetric.MeterProvider
	var lp *sdklog.LoggerProvider

	// Create trace exporter and provider
	traceExporter, err := createTraceExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
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

	// Create batch processor with span filtering
	batchProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)
	// We ignore these spans because they create unneeded noise and extra data we don't need
	ignoredSpans := []string{
		"gocloud.dev/pubsub.driver.Subscription.ReceiveBatch",
	}
	filteredProcessor := newSpanFilterProcessor(batchProcessor, ignoredSpans)

	// Create trace provider
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(filteredProcessor),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sampler)),
	)

	// Set as global trace provider
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

	// Create metrics provider if enabled
	if cfg.EnableMetrics {
		metricExporter, err := createMetricExporter(ctx, cfg)
		if err != nil {
			// Shutdown trace provider before returning error
			_ = tp.Shutdown(ctx)
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}

		// Parse metrics interval
		metricsInterval := 15 * time.Second
		if cfg.MetricsInterval != "" {
			if parsed, err := time.ParseDuration(cfg.MetricsInterval); err == nil {
				metricsInterval = parsed
			}
		}

		mp = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(metricsInterval))),
		)

		// Set as global meter provider
		otel.SetMeterProvider(mp)

		log.Info().
			Str("exporter", cfg.Exporter).
			Str("interval", metricsInterval.String()).
			Msg("OpenTelemetry metrics initialized")
	}

	// Create logging provider if enabled
	if cfg.EnableLogging {
		logExporter, err := createLogExporter(ctx, cfg)
		if err != nil {
			// Shutdown previous providers before returning error
			_ = tp.Shutdown(ctx)
			if mp != nil {
				_ = mp.Shutdown(ctx)
			}
			return nil, fmt.Errorf("failed to create log exporter: %w", err)
		}

		lp = sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		)

		// Set as global logger provider
		global.SetLoggerProvider(lp)

		log.Info().
			Str("exporter", cfg.Exporter).
			Msg("OpenTelemetry logging initialized")
	}

	return &Provider{
		tp:  tp,
		mp:  mp,
		lp:  lp,
		cfg: cfg,
		shutdown: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			var errs []error

			// Shutdown logging provider first
			if lp != nil {
				if err := lp.Shutdown(ctx); err != nil {
					errs = append(errs, fmt.Errorf("log provider shutdown: %w", err))
				}
			}

			// Shutdown metrics provider
			if mp != nil {
				if err := mp.Shutdown(ctx); err != nil {
					errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
				}
			}

			// Shutdown trace provider last
			if err := tp.Shutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("trace provider shutdown: %w", err))
			}

			if len(errs) > 0 {
				return fmt.Errorf("shutdown errors: %v", errs)
			}
			return nil
		},
	}, nil
}

// createTraceExporter creates the appropriate trace exporter based on configuration.
func createTraceExporter(ctx context.Context, cfg *config.OTelConfig) (sdktrace.SpanExporter, error) {
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
			opts = append(opts, otlptracehttp.WithURLPath(cfg.PathPrefix+"/v1/traces"))
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

// createMetricExporter creates the appropriate metric exporter based on configuration.
func createMetricExporter(ctx context.Context, cfg *config.OTelConfig) (sdkmetric.Exporter, error) {
	switch strings.ToLower(cfg.Exporter) {
	case "stdout":
		return stdoutmetric.New(stdoutmetric.WithPrettyPrint())

	case "none", "":
		return &noopMetricExporter{}, nil

	case "otlp":
		return createOTLPMetricExporter(ctx, cfg)

	default:
		return nil, fmt.Errorf("unsupported metric exporter type: %s", cfg.Exporter)
	}
}

// createOTLPMetricExporter creates an OTLP metric exporter based on the protocol configuration.
func createOTLPMetricExporter(ctx context.Context, cfg *config.OTelConfig) (sdkmetric.Exporter, error) {
	headers := parseHeaders(cfg.Headers)

	switch strings.ToLower(cfg.Protocol) {
	case "http":
		opts := []otlpmetrichttp.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlpmetrichttp.WithEndpoint(cfg.Endpoint))
			opts = append(opts, otlpmetrichttp.WithURLPath(cfg.PathPrefix+"/v1/metrics"))
		}
		if cfg.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		if len(headers) > 0 {
			opts = append(opts, otlpmetrichttp.WithHeaders(headers))
		}
		return otlpmetrichttp.New(ctx, opts...)

	case "grpc", "":
		opts := []otlpmetricgrpc.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlpmetricgrpc.WithEndpoint(cfg.Endpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		if len(headers) > 0 {
			opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
		}
		return otlpmetricgrpc.New(ctx, opts...)

	default:
		return nil, fmt.Errorf("unsupported OTLP metric protocol: %s", cfg.Protocol)
	}
}

// createLogExporter creates the appropriate log exporter based on configuration.
func createLogExporter(ctx context.Context, cfg *config.OTelConfig) (sdklog.Exporter, error) {
	switch strings.ToLower(cfg.Exporter) {
	case "stdout":
		return stdoutlog.New(stdoutlog.WithPrettyPrint())

	case "none", "":
		return &noopLogExporter{}, nil

	case "otlp":
		return createOTLPLogExporter(ctx, cfg)

	default:
		return nil, fmt.Errorf("unsupported log exporter type: %s", cfg.Exporter)
	}
}

// createOTLPLogExporter creates an OTLP log exporter based on the protocol configuration.
func createOTLPLogExporter(ctx context.Context, cfg *config.OTelConfig) (sdklog.Exporter, error) {
	headers := parseHeaders(cfg.Headers)

	switch strings.ToLower(cfg.Protocol) {
	case "http":
		opts := []otlploghttp.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlploghttp.WithEndpoint(cfg.Endpoint))
			opts = append(opts, otlploghttp.WithURLPath(cfg.PathPrefix+"/v1/logs"))
		}
		if cfg.Insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		if len(headers) > 0 {
			opts = append(opts, otlploghttp.WithHeaders(headers))
		}
		return otlploghttp.New(ctx, opts...)

	case "grpc", "":
		opts := []otlploggrpc.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlploggrpc.WithEndpoint(cfg.Endpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		if len(headers) > 0 {
			opts = append(opts, otlploggrpc.WithHeaders(headers))
		}
		return otlploggrpc.New(ctx, opts...)

	default:
		return nil, fmt.Errorf("unsupported OTLP log protocol: %s", cfg.Protocol)
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

// noopMetricExporter is a no-op metric exporter for when metrics are disabled.
type noopMetricExporter struct{}

func (e *noopMetricExporter) Temporality(_ sdkmetric.InstrumentKind) metricdata.Temporality {
	// Use cumulative temporality as a reasonable default for no-op exporter
	return metricdata.CumulativeTemporality
}

func (e *noopMetricExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	// Use the SDK default aggregation for the given instrument kind
	return sdkmetric.DefaultAggregationSelector(kind)
}

func (e *noopMetricExporter) Export(_ context.Context, _ *metricdata.ResourceMetrics) error {
	return nil
}

func (e *noopMetricExporter) ForceFlush(_ context.Context) error {
	return nil
}

func (e *noopMetricExporter) Shutdown(_ context.Context) error {
	return nil
}

// noopLogExporter is a no-op log exporter for when logging is disabled.
type noopLogExporter struct{}

func (e *noopLogExporter) Export(_ context.Context, _ []sdklog.Record) error {
	return nil
}

func (e *noopLogExporter) ForceFlush(_ context.Context) error {
	return nil
}

func (e *noopLogExporter) Shutdown(_ context.Context) error {
	return nil
}

// spanFilterProcessor wraps a SpanProcessor and filters spans based on their name.
type spanFilterProcessor struct {
	next         sdktrace.SpanProcessor
	ignoredSpans map[string]bool
}

// newSpanFilterProcessor creates a new span filter processor that wraps another processor.
func newSpanFilterProcessor(next sdktrace.SpanProcessor, ignoredSpans []string) sdktrace.SpanProcessor {
	ignored := make(map[string]bool, len(ignoredSpans))
	for _, span := range ignoredSpans {
		ignored[span] = true
	}
	return &spanFilterProcessor{
		next:         next,
		ignoredSpans: ignored,
	}
}

func (s *spanFilterProcessor) OnStart(parent context.Context, span sdktrace.ReadWriteSpan) {
	if !s.ignoredSpans[span.Name()] {
		s.next.OnStart(parent, span)
	}
}

func (s *spanFilterProcessor) OnEnd(span sdktrace.ReadOnlySpan) {
	if !s.ignoredSpans[span.Name()] {
		s.next.OnEnd(span)
	}
}

func (s *spanFilterProcessor) Shutdown(ctx context.Context) error {
	return s.next.Shutdown(ctx)
}

func (s *spanFilterProcessor) ForceFlush(ctx context.Context) error {
	return s.next.ForceFlush(ctx)
}

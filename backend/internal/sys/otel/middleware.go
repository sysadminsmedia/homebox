package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware returns an HTTP middleware that instruments requests with OpenTelemetry.
// If the provider is nil or tracing is disabled, returns a no-op middleware.
func (p *Provider) HTTPMiddleware(operation string) func(http.Handler) http.Handler {
	if p == nil || !p.IsEnabled() || (p.cfg != nil && !p.cfg.EnableHTTPTracing) {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, operation,
			otelhttp.WithTracerProvider(p.TracerProvider()),
			otelhttp.WithSpanNameFormatter(spanNameFormatter),
		)
	}
}

// spanNameFormatter formats the span name to include the HTTP method and route.
func spanNameFormatter(_ string, r *http.Request) string {
	return r.Method + " " + r.URL.Path
}

// WrapHandler wraps an http.Handler with OpenTelemetry instrumentation.
func (p *Provider) WrapHandler(h http.Handler, operation string) http.Handler {
	if p == nil || !p.IsEnabled() || (p.cfg != nil && !p.cfg.EnableHTTPTracing) {
		return h
	}

	return otelhttp.NewHandler(h, operation,
		otelhttp.WithTracerProvider(p.TracerProvider()),
		otelhttp.WithSpanNameFormatter(spanNameFormatter),
	)
}

// SpanFromRequest extracts the current span from an HTTP request context.
func SpanFromRequest(r *http.Request) trace.Span {
	return trace.SpanFromContext(r.Context())
}

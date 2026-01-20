package otel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// FrontendSpan represents a span sent from the frontend.
type FrontendSpan struct {
	TraceID      string                 `json:"traceId"`
	SpanID       string                 `json:"spanId"`
	ParentSpanID string                 `json:"parentSpanId,omitempty"`
	Name         string                 `json:"name"`
	Kind         string                 `json:"kind,omitempty"`
	StartTime    int64                  `json:"startTime"` // Unix milliseconds
	EndTime      int64                  `json:"endTime"`   // Unix milliseconds
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
	Status       *SpanStatus            `json:"status,omitempty"`
	Events       []SpanEvent            `json:"events,omitempty"`
}

// SpanStatus represents the status of a span.
type SpanStatus struct {
	Code    string `json:"code"` // "ok", "error", "unset"
	Message string `json:"message,omitempty"`
}

// SpanEvent represents an event within a span.
type SpanEvent struct {
	Name       string                 `json:"name"`
	Time       int64                  `json:"time"` // Unix milliseconds
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// TelemetryPayload represents the payload sent from the frontend.
type TelemetryPayload struct {
	ResourceAttributes map[string]interface{} `json:"resourceAttributes,omitempty"`
	Spans              []FrontendSpan         `json:"spans"`
}

// MaxTelemetryPayloadSize is the maximum size of telemetry payloads accepted by the proxy.
// This prevents denial-of-service attacks through excessively large payloads.
const MaxTelemetryPayloadSize = 1024 * 1024 // 1MB

// ProxyHandler handles telemetry data from the frontend.
// This creates new spans in the backend that represent frontend activity,
// allowing for distributed tracing across the full stack.
func (p *Provider) ProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p == nil || !p.IsEnabled() || p.cfg == nil || !p.cfg.ProxyEnabled {
			http.Error(w, "Telemetry proxy is disabled", http.StatusServiceUnavailable)
			return
		}

		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read and parse the payload with size limit
		body, err := io.ReadAll(io.LimitReader(r.Body, MaxTelemetryPayloadSize))
		if err != nil {
			log.Error().Err(err).Msg("failed to read telemetry payload")
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var payload TelemetryPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Error().Err(err).Msg("failed to parse telemetry payload")
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Process the spans
		processed := 0
		for _, span := range payload.Spans {
			if err := p.processSpan(r.Context(), span, payload.ResourceAttributes); err != nil {
				log.Warn().Err(err).Str("span", span.Name).Msg("failed to process span")
				continue
			}
			processed++
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"processed": processed,
			"total":     len(payload.Spans),
		})
	}
}

// processSpan processes a single frontend span and creates a corresponding backend span.
func (p *Provider) processSpan(ctx context.Context, span FrontendSpan, resourceAttrs map[string]interface{}) error {
	tracer := p.Tracer("frontend")

	// Parse span kind
	spanKind := trace.SpanKindInternal
	switch strings.ToLower(span.Kind) {
	case "client":
		spanKind = trace.SpanKindClient
	case "server":
		spanKind = trace.SpanKindServer
	case "producer":
		spanKind = trace.SpanKindProducer
	case "consumer":
		spanKind = trace.SpanKindConsumer
	}

	// Convert timestamps
	startTime := time.UnixMilli(span.StartTime)
	endTime := time.UnixMilli(span.EndTime)

	// Build attributes
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.String("span.origin", "frontend"))
	attrs = append(attrs, attribute.String("frontend.trace_id", span.TraceID))
	attrs = append(attrs, attribute.String("frontend.span_id", span.SpanID))
	if span.ParentSpanID != "" {
		attrs = append(attrs, attribute.String("frontend.parent_span_id", span.ParentSpanID))
	}

	// Add span attributes
	for k, v := range span.Attributes {
		attrs = append(attrs, attributeFromInterface("frontend."+k, v))
	}

	// Add resource attributes
	for k, v := range resourceAttrs {
		attrs = append(attrs, attributeFromInterface("frontend.resource."+k, v))
	}

	// Create the span
	_, backendSpan := tracer.Start(ctx, "frontend: "+span.Name,
		trace.WithSpanKind(spanKind),
		trace.WithTimestamp(startTime),
		trace.WithAttributes(attrs...),
	)

	// Add events
	for _, event := range span.Events {
		eventAttrs := make([]attribute.KeyValue, 0)
		for k, v := range event.Attributes {
			eventAttrs = append(eventAttrs, attributeFromInterface(k, v))
		}
		backendSpan.AddEvent(event.Name,
			trace.WithTimestamp(time.UnixMilli(event.Time)),
			trace.WithAttributes(eventAttrs...),
		)
	}

	// Set status
	if span.Status != nil {
		switch strings.ToLower(span.Status.Code) {
		case "error":
			backendSpan.SetStatus(codes.Error, span.Status.Message)
		case "ok":
			backendSpan.SetStatus(codes.Ok, span.Status.Message)
		}
	}

	// End the span with the correct end time
	backendSpan.End(trace.WithTimestamp(endTime))

	return nil
}

// attributeFromInterface converts an interface value to an OpenTelemetry attribute.
func attributeFromInterface(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int:
		return attribute.IntSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

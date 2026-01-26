package otel

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	collogspb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
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
// This forwards frontend telemetry data to the configured OTLP endpoint,
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
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Error().Err(err).Msg("failed to close telemetry payload body")
			}
		}(r.Body)

		var payload TelemetryPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Error().Err(err).Msg("failed to parse telemetry payload")
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Forward the spans to the OTLP endpoint
		if err := p.forwardToOTLP(r.Context(), payload); err != nil {
			log.Error().Err(err).Msg("failed to forward telemetry to OTLP endpoint")
			http.Error(w, "Failed to forward telemetry", http.StatusInternalServerError)
			return
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"processed": len(payload.Spans),
			"total":     len(payload.Spans),
		})
	}
}

// forwardToOTLP converts frontend spans to OTLP format and forwards them to the configured endpoint.
func (p *Provider) forwardToOTLP(ctx context.Context, payload TelemetryPayload) error {
	// Create an OTLP client based on configuration
	client, err := p.createProxyClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP client: %w", err)
	}
	defer func() {
		if shutdownErr := client.Shutdown(ctx); shutdownErr != nil {
			log.Warn().Err(shutdownErr).Msg("failed to shutdown OTLP client")
		}
	}()

	// Convert frontend spans to OTLP protobuf spans
	otlpSpans, err := p.convertToOTLPSpans(payload)
	if err != nil {
		return fmt.Errorf("failed to convert spans to OTLP format: %w", err)
	}

	// Upload the spans to the OTLP endpoint
	if err := client.UploadTraces(ctx, otlpSpans); err != nil {
		return fmt.Errorf("failed to upload traces: %w", err)
	}

	return nil
}

// otlpClient provides a unified interface for OTLP trace uploading
type otlpClient interface {
	UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error
	Shutdown(ctx context.Context) error
}

// grpcClient wraps the gRPC OTLP client
type grpcClient struct {
	client collogspb.TraceServiceClient
	conn   *grpc.ClientConn
}

func (c *grpcClient) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	req := &collogspb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	}
	_, err := c.client.Export(ctx, req)
	return err
}

func (c *grpcClient) Shutdown(_ context.Context) error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// httpClient wraps the HTTP OTLP client
type httpClient struct {
	client  *http.Client
	url     string
	headers map[string]string
}

func (c *httpClient) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	req := &collogspb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	}

	// Marshal to protobuf
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	// Send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *httpClient) Shutdown(_ context.Context) error {
	// HTTP client doesn't need explicit shutdown
	return nil
}

// createProxyClient creates an OTLP client for forwarding frontend spans.
func (p *Provider) createProxyClient(_ context.Context) (otlpClient, error) {
	headers := parseHeaders(p.cfg.Headers)

	switch strings.ToLower(p.cfg.Protocol) {
	case "http":
		scheme := "https"
		if p.cfg.Insecure {
			scheme = "http"
		}
		url := fmt.Sprintf("%s://%s/v1/traces", scheme, p.cfg.Endpoint)

		return &httpClient{
			client:  &http.Client{Timeout: 30 * time.Second},
			url:     url,
			headers: headers,
		}, nil

	case "grpc", "":
		dialOpts := []grpc.DialOption{}
		if p.cfg.Insecure {
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		conn, err := grpc.NewClient(p.cfg.Endpoint, dialOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC client: %w", err)
		}

		client := collogspb.NewTraceServiceClient(conn)
		return &grpcClient{
			client: client,
			conn:   conn,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported OTLP protocol: %s", p.cfg.Protocol)
	}
}

// convertToOTLPSpans converts frontend spans to OTLP protobuf format
func (p *Provider) convertToOTLPSpans(payload TelemetryPayload) ([]*tracepb.ResourceSpans, error) {
	// Group spans by resource (in our case, all frontend spans share the same resource)
	serviceName := p.cfg.ServiceName + "-frontend"

	resourceAttrs := make([]*commonpb.KeyValue, 0)
	resourceAttrs = append(resourceAttrs, &commonpb.KeyValue{
		Key:   "service.name",
		Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: serviceName}},
	})

	for k, v := range payload.ResourceAttributes {
		resourceAttrs = append(resourceAttrs, convertAttributeToProto(k, v))
	}

	// Convert all spans
	otlpSpans := make([]*tracepb.Span, 0, len(payload.Spans))
	for _, frontendSpan := range payload.Spans {
		otlpSpan, err := convertSpanToProto(frontendSpan)
		if err != nil {
			log.Warn().Err(err).Str("span", frontendSpan.Name).Msg("failed to convert span, skipping")
			continue
		}
		otlpSpans = append(otlpSpans, otlpSpan)
	}

	if len(otlpSpans) == 0 {
		return nil, fmt.Errorf("no valid spans to export")
	}

	// Create the resource spans structure
	resourceSpans := []*tracepb.ResourceSpans{
		{
			Resource: &resourcepb.Resource{
				Attributes: resourceAttrs,
			},
			ScopeSpans: []*tracepb.ScopeSpans{
				{
					Scope: &commonpb.InstrumentationScope{
						Name:    serviceName,
						Version: p.cfg.ServiceVersion,
					},
					Spans: otlpSpans,
				},
			},
		},
	}

	return resourceSpans, nil
}

// convertSpanToProto converts a single frontend span to OTLP protobuf format
func convertSpanToProto(frontendSpan FrontendSpan) (*tracepb.Span, error) {
	// Parse trace ID
	traceID, err := hex.DecodeString(frontendSpan.TraceID)
	if err != nil {
		return nil, fmt.Errorf("invalid trace ID: %w", err)
	}

	// Parse span ID
	spanID, err := hex.DecodeString(frontendSpan.SpanID)
	if err != nil {
		return nil, fmt.Errorf("invalid span ID: %w", err)
	}

	// Parse parent span ID if present
	var parentSpanID []byte
	if frontendSpan.ParentSpanID != "" {
		parentSpanID, err = hex.DecodeString(frontendSpan.ParentSpanID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent span ID: %w", err)
		}
	}

	// Convert span kind
	spanKind := tracepb.Span_SPAN_KIND_INTERNAL
	switch strings.ToLower(frontendSpan.Kind) {
	case "client":
		spanKind = tracepb.Span_SPAN_KIND_CLIENT
	case "server":
		spanKind = tracepb.Span_SPAN_KIND_SERVER
	case "producer":
		spanKind = tracepb.Span_SPAN_KIND_PRODUCER
	case "consumer":
		spanKind = tracepb.Span_SPAN_KIND_CONSUMER
	}

	// Convert attributes
	attrs := make([]*commonpb.KeyValue, 0, len(frontendSpan.Attributes))
	for k, v := range frontendSpan.Attributes {
		attrs = append(attrs, convertAttributeToProto(k, v))
	}

	// Convert events
	events := make([]*tracepb.Span_Event, 0, len(frontendSpan.Events))
	for _, event := range frontendSpan.Events {
		eventAttrs := make([]*commonpb.KeyValue, 0, len(event.Attributes))
		for k, v := range event.Attributes {
			eventAttrs = append(eventAttrs, convertAttributeToProto(k, v))
		}
		events = append(events, &tracepb.Span_Event{
			TimeUnixNano: uint64(event.Time * 1000000), // Convert milliseconds to nanoseconds
			Name:         event.Name,
			Attributes:   eventAttrs,
		})
	}

	// Convert status
	status := &tracepb.Status{
		Code: tracepb.Status_STATUS_CODE_UNSET,
	}
	if frontendSpan.Status != nil {
		switch strings.ToLower(frontendSpan.Status.Code) {
		case "ok":
			status.Code = tracepb.Status_STATUS_CODE_OK
			status.Message = frontendSpan.Status.Message
		case "error":
			status.Code = tracepb.Status_STATUS_CODE_ERROR
			status.Message = frontendSpan.Status.Message
		}
	}

	return &tracepb.Span{
		TraceId:           traceID,
		SpanId:            spanID,
		ParentSpanId:      parentSpanID,
		Name:              frontendSpan.Name,
		Kind:              spanKind,
		StartTimeUnixNano: uint64(frontendSpan.StartTime * 1000000), // Convert milliseconds to nanoseconds
		EndTimeUnixNano:   uint64(frontendSpan.EndTime * 1000000),   // Convert milliseconds to nanoseconds
		Attributes:        attrs,
		Events:            events,
		Status:            status,
	}, nil
}

// convertAttributeToProto converts an attribute to OTLP protobuf format
func convertAttributeToProto(key string, value interface{}) *commonpb.KeyValue {
	var anyValue *commonpb.AnyValue

	switch v := value.(type) {
	case string:
		anyValue = &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: v}}
	case int:
		anyValue = &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(v)}}
	case int64:
		anyValue = &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: v}}
	case float64:
		anyValue = &commonpb.AnyValue{Value: &commonpb.AnyValue_DoubleValue{DoubleValue: v}}
	case bool:
		anyValue = &commonpb.AnyValue{Value: &commonpb.AnyValue_BoolValue{BoolValue: v}}
	default:
		anyValue = &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: fmt.Sprintf("%v", v)}}
	}

	return &commonpb.KeyValue{
		Key:   key,
		Value: anyValue,
	}
}

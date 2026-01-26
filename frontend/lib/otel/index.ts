/**
 * OpenTelemetry configuration and initialization for the frontend.
 * This module sets up browser-based tracing that connects to backend traces.
 */

import { WebTracerProvider, BatchSpanProcessor } from "@opentelemetry/sdk-trace-web";
import type { SpanExporter, ReadableSpan } from "@opentelemetry/sdk-trace-web";
import { ZoneContextManager } from "@opentelemetry/context-zone";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { FetchInstrumentation } from "@opentelemetry/instrumentation-fetch";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { ATTR_SERVICE_NAME, ATTR_SERVICE_VERSION } from "@opentelemetry/semantic-conventions";
import type { Span, Attributes } from "@opentelemetry/api";
import { trace, context, propagation, SpanKind, SpanStatusCode } from "@opentelemetry/api";
import { ExportResultCode } from "@opentelemetry/core";
import type { ExportResult } from "@opentelemetry/core";

// Types for the telemetry configuration
export interface OTelConfig {
  enabled: boolean;
  serviceName: string;
  serviceVersion: string;
  // Always use backend proxy for frontend telemetry (required for authentication)
  useBackendProxy: boolean;
  // Sampling rate (0.0 to 1.0)
  sampleRate: number;
  // Enable console logging for debugging
  debug: boolean;
}

// Default configuration
const defaultConfig: OTelConfig = {
  enabled: false,
  serviceName: "homebox-frontend",
  serviceVersion: "1.0.0",
  useBackendProxy: true,
  sampleRate: 1.0,
  debug: false,
};

// Global provider instance
let provider: WebTracerProvider | null = null;
let isInitialized = false;

/**
 * Custom span exporter that sends spans to the backend proxy endpoint.
 */
class BackendProxyExporter implements SpanExporter {
  private endpoint: string;
  private debug: boolean;

  constructor(endpoint: string, debug: boolean = false) {
    this.endpoint = endpoint;
    this.debug = debug;
  }

  export(spans: ReadableSpan[], resultCallback: (result: ExportResult) => void): void {
    if (spans.length === 0) {
      resultCallback({ code: ExportResultCode.SUCCESS });
      return;
    }

    const payload = {
      resourceAttributes: {},
      spans: spans.map(span => this.convertSpan(span)),
    };

    if (this.debug) {
      console.log("[OTel] Exporting spans:", payload);
    }

    fetch(this.endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
      credentials: "include", // Include auth cookies
    })
      .then(response => {
        if (response.ok) {
          resultCallback({ code: ExportResultCode.SUCCESS });
        } else {
          console.warn("[OTel] Failed to export spans:", response.status);
          resultCallback({ code: ExportResultCode.FAILED });
        }
      })
      .catch(error => {
        console.warn("[OTel] Error exporting spans:", error);
        resultCallback({ code: ExportResultCode.FAILED });
      });
  }

  shutdown(): Promise<void> {
    return Promise.resolve();
  }

  forceFlush(): Promise<void> {
    return Promise.resolve();
  }

  private convertSpan(span: ReadableSpan): object {
    const spanContext = span.spanContext();
    // Get parent span ID from the span context if available
    const parentSpanContext = (span as unknown as { parentSpanId?: string }).parentSpanId;
    return {
      traceId: spanContext.traceId,
      spanId: spanContext.spanId,
      parentSpanId: parentSpanContext || undefined,
      name: span.name,
      kind: this.spanKindToString(span.kind),
      startTime: span.startTime[0] * 1000 + Math.floor(span.startTime[1] / 1000000),
      endTime: span.endTime[0] * 1000 + Math.floor(span.endTime[1] / 1000000),
      attributes: this.convertAttributes(span.attributes),
      status:
        span.status.code !== SpanStatusCode.UNSET
          ? {
              code: span.status.code === SpanStatusCode.OK ? "ok" : "error",
              message: span.status.message || undefined,
            }
          : undefined,
      events: span.events.map(event => ({
        name: event.name,
        time: event.time[0] * 1000 + Math.floor(event.time[1] / 1000000),
        attributes: this.convertAttributes(event.attributes || {}),
      })),
    };
  }

  private spanKindToString(kind: SpanKind): string {
    switch (kind) {
      case SpanKind.CLIENT:
        return "client";
      case SpanKind.SERVER:
        return "server";
      case SpanKind.PRODUCER:
        return "producer";
      case SpanKind.CONSUMER:
        return "consumer";
      default:
        return "internal";
    }
  }

  private convertAttributes(attributes: Attributes): Record<string, unknown> {
    const result: Record<string, unknown> = {};
    if (attributes) {
      for (const [key, value] of Object.entries(attributes)) {
        result[key] = value;
      }
    }
    return result;
  }
}

/**
 * Initialize OpenTelemetry tracing for the frontend.
 * Should be called once during app startup.
 */
export function initializeOTel(config: Partial<OTelConfig> = {}): void {
  if (isInitialized) {
    console.warn("[OTel] Already initialized");
    return;
  }

  const finalConfig = { ...defaultConfig, ...config };

  if (!finalConfig.enabled) {
    if (finalConfig.debug) {
      console.log("[OTel] Tracing is disabled");
    }
    return;
  }

  // Create resource with service information
  const resource = resourceFromAttributes({
    [ATTR_SERVICE_NAME]: finalConfig.serviceName,
    [ATTR_SERVICE_VERSION]: finalConfig.serviceVersion,
  });

  // Configure exporter - always use backend proxy for authentication
  const exporter: SpanExporter = new BackendProxyExporter("/api/v1/telemetry", finalConfig.debug);

  // Use batch processor for better performance
  const batchProcessor = new BatchSpanProcessor(exporter);

  // Create the trace provider with span processor
  provider = new WebTracerProvider({
    resource,
    spanProcessors: [batchProcessor],
  });

  // Register the provider globally
  provider.register({
    contextManager: new ZoneContextManager(),
  });

  // Register auto-instrumentations
  registerInstrumentations({
    instrumentations: [
      new FetchInstrumentation({
        // Only propagate trace headers to same-origin requests to prevent leaking
        // trace information to external domains. The pattern matches the current origin.
        propagateTraceHeaderCorsUrls: [new RegExp(`^${window.location.origin}`)],
        clearTimingResources: true,
        applyCustomAttributesOnSpan: (span, request) => {
          // Add custom attributes to fetch spans
          if (request instanceof Request) {
            span.setAttribute("http.url", request.url);
          }
        },
      }),
    ],
  });

  isInitialized = true;

  if (finalConfig.debug) {
    console.log("[OTel] Tracing initialized", finalConfig);
  }
}

/**
 * Get a tracer for creating custom spans.
 */
export function getTracer(name: string = "homebox-frontend") {
  return trace.getTracer(name);
}

/**
 * Create a span for a custom operation.
 * Returns a function to end the span.
 */
export function startSpan(
  name: string,
  options: {
    kind?: SpanKind;
    attributes?: Record<string, string | number | boolean>;
  } = {}
): { span: Span; end: () => void } {
  const tracer = getTracer();
  const span = tracer.startSpan(name, {
    kind: options.kind || SpanKind.INTERNAL,
    attributes: options.attributes,
  });

  return {
    span,
    end: () => span.end(),
  };
}

/**
 * Execute a function within a span context.
 */
export async function withSpan<T>(
  name: string,
  fn: (span: Span) => Promise<T> | T,
  options: {
    kind?: SpanKind;
    attributes?: Record<string, string | number | boolean>;
  } = {}
): Promise<T> {
  const tracer = getTracer();
  return tracer.startActiveSpan(name, { kind: options.kind || SpanKind.INTERNAL }, async span => {
    if (options.attributes) {
      for (const [key, value] of Object.entries(options.attributes)) {
        span.setAttribute(key, value);
      }
    }
    try {
      const result = await fn(span);
      span.setStatus({ code: SpanStatusCode.OK });
      return result;
    } catch (error) {
      span.setStatus({
        code: SpanStatusCode.ERROR,
        message: error instanceof Error ? error.message : String(error),
      });
      throw error;
    } finally {
      span.end();
    }
  });
}

/**
 * Add the trace context headers to a request.
 * Use this when making fetch requests manually.
 */
export function injectTraceHeaders(headers: Headers | Record<string, string>): void {
  const carrier: Record<string, string> = {};
  propagation.inject(context.active(), carrier);

  if (headers instanceof Headers) {
    for (const [key, value] of Object.entries(carrier)) {
      headers.set(key, value);
    }
  } else {
    Object.assign(headers, carrier);
  }
}

/**
 * Get the current trace context for logging or debugging.
 */
export function getCurrentTraceContext(): { traceId: string; spanId: string } | null {
  const span = trace.getActiveSpan();
  if (!span) {
    return null;
  }
  const ctx = span.spanContext();
  return {
    traceId: ctx.traceId,
    spanId: ctx.spanId,
  };
}

/**
 * Check if OpenTelemetry is initialized.
 */
export function isOTelInitialized(): boolean {
  return isInitialized;
}

/**
 * Shutdown the OpenTelemetry provider.
 * Call this during app cleanup.
 */
export async function shutdownOTel(): Promise<void> {
  if (provider) {
    await provider.shutdown();
    provider = null;
    isInitialized = false;
  }
}

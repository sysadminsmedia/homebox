/**
 * Composable for using OpenTelemetry tracing in Vue components.
 * Provides convenience methods for creating spans and tracking operations.
 */

import { withSpan, startSpan, getCurrentTraceContext, isOTelInitialized } from "~~/lib/otel";
import { SpanKind } from "@opentelemetry/api";
import type { Span } from "@opentelemetry/api";

export function useOTel() {
  const nuxtApp = useNuxtApp();
  const isEnabled = computed(() => nuxtApp.$otelEnabled || false);

  /**
   * Execute an async function within a traced span.
   * The span will automatically be ended when the function completes.
   */
  async function traced<T>(
    name: string,
    fn: (span: Span) => Promise<T> | T,
    options?: {
      kind?: SpanKind;
      attributes?: Record<string, string | number | boolean>;
    }
  ): Promise<T> {
    if (!isEnabled.value) {
      // If OTel is disabled, just run the function
      return fn({} as Span);
    }
    return withSpan(name, fn, options);
  }

  /**
   * Start a manual span. Remember to call end() when done.
   */
  function trace(
    name: string,
    options?: {
      kind?: SpanKind;
      attributes?: Record<string, string | number | boolean>;
    }
  ): { span: Span; end: () => void } {
    if (!isEnabled.value) {
      // Return a no-op span if OTel is disabled
      return {
        span: {} as Span,
        end: () => {},
      };
    }
    return startSpan(name, options);
  }

  /**
   * Get the current trace context (useful for logging correlation).
   */
  function getTraceContext() {
    if (!isEnabled.value) {
      return null;
    }
    return getCurrentTraceContext();
  }

  /**
   * Trace a page navigation.
   */
  function tracePage(pageName: string, attributes?: Record<string, string | number | boolean>) {
    if (!isEnabled.value) {
      return { end: () => {} };
    }

    return trace(`page:${pageName}`, {
      kind: SpanKind.INTERNAL,
      attributes: {
        "page.name": pageName,
        ...attributes,
      },
    });
  }

  /**
   * Trace a user interaction.
   */
  async function traceInteraction<T>(
    actionName: string,
    fn: () => Promise<T> | T,
    attributes?: Record<string, string | number | boolean>
  ): Promise<T> {
    return traced(`interaction:${actionName}`, () => fn(), {
      kind: SpanKind.INTERNAL,
      attributes: {
        "interaction.name": actionName,
        ...attributes,
      },
    });
  }

  return {
    isEnabled,
    traced,
    trace,
    getTraceContext,
    tracePage,
    traceInteraction,
    SpanKind,
    isOTelInitialized,
  };
}

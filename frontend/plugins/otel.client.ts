/**
 * Nuxt plugin to initialize OpenTelemetry on the client side.
 * This ensures tracing is set up early in the application lifecycle.
 */

import { initializeOTel } from "~~/lib/otel";

export default defineNuxtPlugin(() => {
  // Only run on client side
  if (import.meta.client) {
    // Check for runtime config for OTel settings
    // These can be set via environment variables or nuxt.config.ts
    const runtimeConfig = useRuntimeConfig();

    // Get OTel configuration from runtime config or use defaults
    const otelEnabled = String(runtimeConfig.public?.otelEnabled || "false");
    const otelDebug = String(runtimeConfig.public?.otelDebug || "false");

    const otelConfig = {
      enabled: otelEnabled === "true",
      serviceName: String(runtimeConfig.public?.otelServiceName || "homebox-frontend"),
      serviceVersion: String(runtimeConfig.public?.otelServiceVersion || "1.0.0"),
      useBackendProxy: true, // Always use backend proxy for security
      sampleRate: parseFloat(String(runtimeConfig.public?.otelSampleRate || "1.0")),
      debug: otelDebug === "true",
    };

    // Initialize OpenTelemetry
    initializeOTel(otelConfig);

    // Provide composable for accessing OTel in components
    return {
      provide: {
        otelEnabled: otelConfig.enabled,
      },
    };
  }
});

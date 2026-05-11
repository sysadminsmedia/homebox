/**
 * Integration adapter registry for external link services (Paperless, Immich, …).
 *
 * To add a new service:
 *  1. Add an `extractXyzId` function below using `extractWithPattern`.
 *  2. Push a new `ServiceAdapter` entry to `SERVICE_ADAPTERS`.
 *  That's it – drop detection, classification, and hydration are all registry-driven.
 *
 * To remove a service (e.g. Immich):
 *  1. Delete its `ServiceAdapter` entry from `SERVICE_ADAPTERS`.
 *  2. Delete its extract function and any service-specific template branch.
 */

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ServiceAdapter {
  /** Unique service identifier. Also used as the MIME type prefix and settings key prefix. */
  name: string;
  /** Full MIME type stored on attachments, e.g. "paperless/document". */
  mimeType: string;
  /** User-settings key for the service base URL. */
  settingsUrlKey: string;
  /** User-settings key for the service API token. */
  settingsTokenKey: string;
  /** Extract the provider-specific ID from a URL. Returns null when not matched. */
  extractId: (url: string, baseUrl?: string) => string | null;
  /** Build a default human-readable title for a newly linked attachment. */
  buildTitle: (id: string) => string;
}

// ---------------------------------------------------------------------------
// Shared extraction helper
// ---------------------------------------------------------------------------

/**
 * Try to extract a capture group from `url`'s path using `pattern`.
 * If `baseUrl` is provided and parseable, the host must match and the path
 * must start at the configured base path. If `baseUrl` is absent or invalid,
 * the pattern is matched against the full pathname (heuristic fallback).
 */
function extractWithPattern(url: string, baseUrl: string | undefined, pattern: RegExp): string | null {
  const trimmedUrl = url.trim();
  if (!trimmedUrl) return null;

  // Normalise bare hostnames (e.g. "localhost/documents/2") to a parseable URL.
  const normalisedUrl = /^https?:\/\//i.test(trimmedUrl) ? trimmedUrl : `http://${trimmedUrl}`;
  try {
    const target = new URL(normalisedUrl);
    let basePath = "";

    if (baseUrl?.trim()) {
      try {
        const base = new URL(baseUrl.trim());
        if (base.origin !== target.origin) return null;
        basePath = base.pathname.replace(/\/$/, "");
        if (basePath && !target.pathname.startsWith(basePath + "/") && target.pathname !== basePath) {
          return null;
        }
      } catch {
        // Invalid configured base URL – fall through to pattern-only match.
      }
    }

    const pathAfterBase = target.pathname.slice(basePath.length || 0);
    return pathAfterBase.match(pattern)?.[1] ?? null;
  } catch {
    return null;
  }
}

// ---------------------------------------------------------------------------
// Per-service ID extractors (one function per service)
// ---------------------------------------------------------------------------

/** Extract Paperless document ID from patterns: /documents/{id} or /documents/{id}/details */
export function extractPaperlessDocId(url: string, baseUrl?: string): string | null {
  return extractWithPattern(url, baseUrl, /\/documents\/(\d+)(?:\/details)?\/?$/);
}

/** Extract Immich asset UUID from pattern: /assets/{uuid} */
export function extractImmichAssetId(url: string, baseUrl?: string): string | null {
  return extractWithPattern(url, baseUrl, /\/assets\/([a-f0-9-]+)\/?$/i);
}

// ---------------------------------------------------------------------------
// Service registry – the single source of truth for all integrations
// ---------------------------------------------------------------------------

export const SERVICE_ADAPTERS: ServiceAdapter[] = [
  {
    name: "paperless",
    mimeType: "paperless/document",
    settingsUrlKey: "paperless_url",
    settingsTokenKey: "paperless_token",
    extractId: extractPaperlessDocId,
    buildTitle: id => `Paperless Document ${id}`,
  },
  {
    name: "immich",
    mimeType: "immich/asset",
    settingsUrlKey: "immich_url",
    settingsTokenKey: "immich_token",
    extractId: extractImmichAssetId,
    buildTitle: id => `Immich Asset ${id}`,
  },
];

// ---------------------------------------------------------------------------
// Generic helpers consumed by the rest of the frontend
// ---------------------------------------------------------------------------

/** Look up an adapter by service name. */
export function getAdapter(name: string): ServiceAdapter | undefined {
  return SERVICE_ADAPTERS.find(a => a.name === name);
}

/** Look up an adapter by MIME type. */
export function getAdapterByMimeType(mimeType: string): ServiceAdapter | undefined {
  return SERVICE_ADAPTERS.find(a => a.mimeType === mimeType);
}

/**
 * Detect which integration service a URL belongs to.
 * Strategy:
 *  1. Exact host+base-path match against each adapter's configured URL in settings.
 *  2. Fallback: URL-pattern match across all adapters (works when settings are missing/invalid).
 *
 * @returns The matching `ServiceAdapter`, or `null` if none matched.
 */
export function detectServiceFromUrl(url: string, settings: Record<string, string>): ServiceAdapter | null {
  const trimmedUrl = url.trim();
  if (!trimmedUrl) return null;

  const normUrl = /^https?:\/\//i.test(trimmedUrl) ? trimmedUrl : `http://${trimmedUrl}`;
  try {
    const target = new URL(normUrl);

    // 1. Configured base URL match (most precise)
    for (const adapter of SERVICE_ADAPTERS) {
      const baseUrl = settings[adapter.settingsUrlKey]?.trim();
      if (!baseUrl) continue;
      try {
        const base = new URL(baseUrl);
        if (base.origin !== target.origin) continue;
        const basePath = base.pathname.replace(/\/$/, "");
        if (!basePath || target.pathname === basePath || target.pathname.startsWith(basePath + "/")) {
          return adapter;
        }
      } catch {
        // Unparseable configured URL – skip.
      }
    }
  } catch {
    return null;
  }

  // 2. Pattern-only fallback – skipped for any service that has a configured base URL
  // (if settings are present but host didn't match, we do NOT fall back to pattern;
  //  this prevents host-prefix spoofing: paperless.local.evil won't match paperless.local).
  for (const adapter of SERVICE_ADAPTERS) {
    if (settings[adapter.settingsUrlKey]) continue; // configured but didn't match – skip
    if (adapter.extractId(url) !== null) return adapter;
  }

  return null;
}

/**
 * Classify a dropped URL into a `{ adapter, id }` pair.
 * Tries the configured base URL first, then falls back to pattern-only extraction.
 *
 * @returns `{ adapter, id }` when recognised, `null` when the URL matches no known service.
 */
export function classifyDroppedUrl(
  url: string,
  settings: Record<string, string>
): { adapter: ServiceAdapter; id: string } | null {
  const adapter = detectServiceFromUrl(url, settings);
  if (!adapter) return null;

  const configuredBase = settings[adapter.settingsUrlKey];
  const id = adapter.extractId(url, configuredBase) ?? adapter.extractId(url);
  if (!id) return null;

  return { adapter, id };
}



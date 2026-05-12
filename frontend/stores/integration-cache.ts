import { defineStore } from "pinia";
import { SERVICE_ADAPTERS, type ServiceAdapter } from "~/lib/integration-adapters";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** Lifecycle state of a single service-typed attachment. */
export type AttachmentFetchState = "loading" | "ok" | "stale" | "error";

// ---------------------------------------------------------------------------
// localStorage helpers (TTL-based, key-prefixed)
// ---------------------------------------------------------------------------

const CACHE_TTL_MS = 30 * 60 * 1000; // 30 minutes
const LS_PREFIX = "homebox:integration:";

interface StoredEntry {
  data: unknown;
  fetchedAt: number;
}

function lsRead(key: string): unknown | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(LS_PREFIX + key);
    if (!raw) return null;
    const entry = JSON.parse(raw) as StoredEntry;
    if (Date.now() - entry.fetchedAt > CACHE_TTL_MS) {
      localStorage.removeItem(LS_PREFIX + key);
      return null;
    }
    return entry.data;
  } catch {
    return null;
  }
}

function lsWrite(key: string, data: unknown): void {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(LS_PREFIX + key, JSON.stringify({ data, fetchedAt: Date.now() } as StoredEntry));
  } catch {
    // localStorage quota exceeded or unavailable – ignore.
  }
}

function lsDelete(key: string): void {
  if (typeof localStorage === "undefined") return;
  localStorage.removeItem(LS_PREFIX + key);
}

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

export const useIntegrationCacheStore = defineStore("integrationCache", () => {
  /**
   * Configured base URL per service name (trailing slash stripped).
   * Empty string = not configured.
   */
  const serviceUrls = reactive<Record<string, string>>({});

  /**
   * Resolved enriched metadata, keyed by `${serviceName}:${id}`.
   * In-memory mirror of the localStorage cache; populated on first read.
   */
  const enrichedData = reactive<Record<string, unknown>>({});

  /** Per-attachment-id fetch state. Key = attachment DB id. */
  const fetchStates = reactive<Record<string, AttachmentFetchState>>({});

  /** Per-attachment-id failure description. Key = attachment DB id. */
  const fetchErrors = reactive<Record<string, string>>({});

  /** Whether settings have been fetched this session (avoid duplicate API calls). */
  const settingsLoaded = ref(false);

  // -------------------------------------------------------------------------
  // Settings
  // -------------------------------------------------------------------------

  /**
   * Load service base URLs from user settings.
   * No-ops on subsequent calls within the same page session.
   */
  async function loadSettings(api: ReturnType<typeof useUserApi>): Promise<void> {
    if (settingsLoaded.value) return;
    const { data, error } = await api.user.getSettings();
    if (error || !data?.item) return;
    const s = data.item as Record<string, unknown>;
    for (const adapter of SERVICE_ADAPTERS) {
      serviceUrls[adapter.name] = ((s[adapter.settingsUrlKey] as string) || "").replace(/\/$/, "");
    }
    settingsLoaded.value = true;
  }

  function isConfigured(adapter: ServiceAdapter): boolean {
    return !!(serviceUrls[adapter.name]?.trim());
  }

  function getUrl(adapter: ServiceAdapter): string {
    return serviceUrls[adapter.name] ?? "";
  }

  // -------------------------------------------------------------------------
  // Enriched data cache
  // -------------------------------------------------------------------------

  function getEnrichedData(serviceName: string, id: string): unknown {
    const key = `${serviceName}:${id}`;
    if (key in enrichedData) return enrichedData[key];
    const cached = lsRead(key);
    if (cached !== null) enrichedData[key] = cached;
    return cached;
  }

  function setEnrichedData(serviceName: string, id: string, data: unknown): void {
    const key = `${serviceName}:${id}`;
    enrichedData[key] = data;
    lsWrite(key, data);
  }

  function invalidateEnrichedData(serviceName: string, id: string): void {
    const key = `${serviceName}:${id}`;
    delete enrichedData[key];
    lsDelete(key);
  }

  // -------------------------------------------------------------------------
  // Attachment fetch state
  // -------------------------------------------------------------------------

  function setState(attachmentId: string, state: AttachmentFetchState, errorMsg?: string): void {
    fetchStates[attachmentId] = state;
    if (errorMsg !== undefined) fetchErrors[attachmentId] = errorMsg;
  }

  function clearAttachmentState(attachmentId: string): void {
    delete fetchStates[attachmentId];
    delete fetchErrors[attachmentId];
  }

  /**
   * Update a single service URL (e.g. after profile save).
   * Resets settingsLoaded so loadSettings() will re-read on next call.
   */
  function setServiceUrl(name: string, url: string): void {
    serviceUrls[name] = url.replace(/\/$/, "");
    settingsLoaded.value = false;
  }

  return {
    serviceUrls,
    enrichedData,
    fetchStates,
    fetchErrors,
    loadSettings,
    isConfigured,
    getUrl,
    setServiceUrl,
    getEnrichedData,
    setEnrichedData,
    invalidateEnrichedData,
    setState,
    clearAttachmentState,
  };
});

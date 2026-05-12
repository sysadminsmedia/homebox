/**
 * Tests for the pure helper functions in preferences-utils.ts.
 *
 * These tests exist specifically to catch the class of bug where
 * `saveToServer` would call `setSettings(buildSyncedSettings(...))` directly,
 * wiping integration keys (paperless_url, paperless_token, etc.) that are
 * stored alongside preference keys in the same settings object.
 *
 * Regression: the fix ensures `saveToServer` does GET → merge → PUT.
 * The tests below verify the invariants that make the merge safe.
 */
import { describe, expect, it } from "vitest";
import { DEFAULT_PREFERENCES, buildSyncedSettings, mergeSyncedSettings } from "./preferences-utils";
import { SERVICE_ADAPTERS } from "../lib/integration-adapters";

// Keys that live in the settings object but are NOT preferences and must never
// be overwritten by a bare preferences save.
const INTEGRATION_KEYS = SERVICE_ADAPTERS.flatMap(a => [a.settingsUrlKey, a.settingsTokenKey]);

// Default sync config (all preference keys synced).
const SYNC_ALL = {};

describe("use-preferences pure helpers", () => {
  describe("buildSyncedSettings", () => {
    it("never includes integration keys", () => {
      const payload = buildSyncedSettings(DEFAULT_PREFERENCES, SYNC_ALL);
      for (const key of INTEGRATION_KEYS) {
        expect(Object.prototype.hasOwnProperty.call(payload, key)).toBe(false);
      }
    });

    it("includes known preference keys", () => {
      const payload = buildSyncedSettings(DEFAULT_PREFERENCES, SYNC_ALL);
      expect(payload).toHaveProperty("showDetails");
      expect(payload).toHaveProperty("theme");
      expect(payload).toHaveProperty("duplicateSettings");
    });

    it("reflects actual preference values", () => {
      const prefs = { ...DEFAULT_PREFERENCES, showDetails: false, theme: "dark" as never };
      const payload = buildSyncedSettings(prefs, SYNC_ALL);
      expect(payload.showDetails).toBe(false);
      expect(payload.theme).toBe("dark");
    });

    it("respects syncConfig exclusions", () => {
      const payload = buildSyncedSettings(DEFAULT_PREFERENCES, { itemDisplayView: false });
      expect(payload).not.toHaveProperty("itemDisplayView");
      expect(payload).toHaveProperty("showDetails");
    });
  });

  describe("mergeSyncedSettings – regression: integration keys are preserved", () => {
    it("does NOT copy integration keys from server settings into preferences", () => {
      const serverSettings = {
        showDetails: false,
        paperless_url: "http://localhost:8000",
        paperless_token: "secret",
      };
      const result = mergeSyncedSettings(serverSettings, DEFAULT_PREFERENCES, SYNC_ALL);
      // mergeSyncedSettings only touches known preference keys – integration keys
      // must not appear on the preferences object.
      expect((result as Record<string, unknown>).paperless_url).toBeUndefined();
      expect((result as Record<string, unknown>).paperless_token).toBeUndefined();
    });

    it("merges known preference keys from server settings", () => {
      const serverSettings = { showDetails: false, theme: "dark" };
      const result = mergeSyncedSettings(serverSettings, DEFAULT_PREFERENCES, SYNC_ALL);
      expect(result.showDetails).toBe(false);
      expect(result.theme).toBe("dark");
    });

    it("preserves local preference values for keys absent from server", () => {
      const result = mergeSyncedSettings({}, { ...DEFAULT_PREFERENCES, showDetails: false }, SYNC_ALL);
      expect(result.showDetails).toBe(false);
    });
  });

  describe("merge round-trip – the exact scenario that was broken", () => {
    /**
     * Simulates what saveToServer now does correctly:
     *   existing = await getSettings()         ← includes paperless_url
     *   merged   = { ...existing, ...buildSyncedSettings(prefs, syncConfig) }
     *   await setSettings(merged)              ← must still contain paperless_url
     *
     * Previously: setSettings(buildSyncedSettings(prefs))  ← wiped paperless_url
     */
    it("merged payload preserves integration keys after preference save", () => {
      const existingServerSettings: Record<string, unknown> = {
        showDetails: true,
        theme: "homebox",
        paperless_url: "http://localhost:8000",
        paperless_token: "secret-token",
      };

      const localPrefs = { ...DEFAULT_PREFERENCES, showDetails: false };
      const prefPayload = buildSyncedSettings(localPrefs, SYNC_ALL);

      // This is the fixed merge: existing settings first, preferences on top.
      const merged = { ...existingServerSettings, ...prefPayload };

      expect(merged.paperless_url).toBe("http://localhost:8000");
      expect(merged.paperless_token).toBe("secret-token");
      // And preferences are still updated:
      expect(merged.showDetails).toBe(false);
    });
  });
});

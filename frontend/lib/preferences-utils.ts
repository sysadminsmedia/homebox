/**
 * Pure, framework-free utilities for view-preference syncing.
 * Extracted so they can be unit-tested without a Nuxt/Vue environment.
 */
import type { EntitySummary } from "~/lib/api/types/data-contracts";
import type { DaisyTheme } from "~~/lib/data/themes";

export type ViewType = "table" | "card";

export type DuplicateSettings = {
  copyMaintenance: boolean;
  copyAttachments: boolean;
  copyCustomFields: boolean;
  copyPrefixOverride: string | null;
};

export type LocationViewPreferences = {
  showDetails: boolean;
  showEmpty: boolean;
  editorAdvancedView: boolean;
  itemDisplayView: ViewType;
  theme: DaisyTheme;
  itemsPerTablePage: number;
  tableHeaders?: {
    value: keyof EntitySummary;
    enabled: boolean;
  }[];
  displayLegacyHeader: boolean;
  legacyImageFit: boolean;
  language?: string | null;
  overrideFormatLocale?: string | null;
  collectionId?: string | null;
  duplicateSettings: DuplicateSettings;
  shownMultiTabWarning: boolean;
  quickActions: {
    enabled: boolean;
  };
};

export type PreferenceSyncConfig = Partial<Record<keyof LocationViewPreferences, boolean>>;

export const DEFAULT_PREFERENCES: LocationViewPreferences = {
  showDetails: true,
  showEmpty: true,
  editorAdvancedView: false,
  itemDisplayView: "card",
  theme: "homebox",
  itemsPerTablePage: 10,
  displayLegacyHeader: false,
  legacyImageFit: false,
  language: null,
  overrideFormatLocale: null,
  duplicateSettings: {
    copyMaintenance: false,
    copyAttachments: true,
    copyCustomFields: true,
    copyPrefixOverride: null,
  },
  shownMultiTabWarning: false,
  quickActions: {
    enabled: true,
  },
};

const preferenceKeys: (keyof LocationViewPreferences)[] = [
  "showDetails",
  "showEmpty",
  "editorAdvancedView",
  "itemDisplayView",
  "theme",
  "itemsPerTablePage",
  "tableHeaders",
  "displayLegacyHeader",
  "legacyImageFit",
  "language",
  "overrideFormatLocale",
  "collectionId",
  "duplicateSettings",
  "shownMultiTabWarning",
  "quickActions",
];

export function forEachSyncedPreference(
  syncConfig: PreferenceSyncConfig,
  callback: (key: keyof LocationViewPreferences) => void
) {
  for (const key of preferenceKeys) {
    if (syncConfig[key] !== false) {
      callback(key);
    }
  }
}

export function buildSyncedSettings(
  preferences: LocationViewPreferences,
  syncConfig: PreferenceSyncConfig
): Record<string, unknown> {
  const payload: Record<string, unknown> = {};
  forEachSyncedPreference(syncConfig, key => {
    payload[key] = preferences[key];
  });
  return payload;
}

export function mergeSyncedSettings(
  settings: Record<string, unknown>,
  preferences: LocationViewPreferences,
  syncConfig: PreferenceSyncConfig
): LocationViewPreferences {
  const nextPreferences = { ...preferences };
  forEachSyncedPreference(syncConfig, key => {
    if (key in settings) {
      // Server settings are schemaless JSON; key selection above limits assignment
      // to known preference fields.
      nextPreferences[key] = settings[key] as never;
    }
  });
  return nextPreferences;
}

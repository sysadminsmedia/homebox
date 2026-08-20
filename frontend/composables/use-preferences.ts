import type { Ref } from "vue";
import type { EntitySummary } from "~/lib/api/types/data-contracts";
import type { DaisyTheme } from "~~/lib/data/themes";
import { resolveWeekStart, type WeekStart } from "~~/lib/datelib/weekStart";

export type ViewType = "table" | "card";

// Re-exported so existing importers of the preference helpers keep a single
// entrypoint; the implementation lives in a side-effect-free module for testing.
export { resolveWeekStart, type WeekStart };

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
  firstDayOfWeek: WeekStart;
  collectionId?: string | null;
  duplicateSettings: DuplicateSettings;
  shownMultiTabWarning: boolean;
  quickActions: {
    enabled: boolean;
  };
};
export type PreferenceSyncConfig = Partial<Record<keyof LocationViewPreferences, boolean>>;
type PreferenceChange = true | Record<string, PreferenceChange>;
type PreferenceChanges = Partial<Record<keyof LocationViewPreferences, PreferenceChange>>;

const DEFAULT_PREFERENCES: LocationViewPreferences = {
  showDetails: true,
  showEmpty: true,
  editorAdvancedView: false,
  itemDisplayView: "card",
  theme: "homebox",
  itemsPerTablePage: 12,
  displayLegacyHeader: false,
  legacyImageFit: false,
  language: null,
  overrideFormatLocale: null,
  firstDayOfWeek: "auto",
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

let syncConfig: PreferenceSyncConfig = {
  itemDisplayView: false,
  shownMultiTabWarning: false,
};

let syncInitialized = false;

const preferenceKeys = Object.keys(DEFAULT_PREFERENCES) as (keyof LocationViewPreferences)[];

const results = useLocalStorage("homebox/preferences/location", DEFAULT_PREFERENCES, { mergeDefaults: true });

function forEachSyncedPreference(callback: (key: keyof LocationViewPreferences) => void) {
  for (const key of preferenceKeys) {
    if (syncConfig[key] !== false) {
      callback(key);
    }
  }
}

function buildSyncedSettings(preferences: LocationViewPreferences): Record<string, unknown> {
  const payload: Record<string, unknown> = {};
  forEachSyncedPreference(key => {
    payload[key] = preferences[key];
  });
  return payload;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function isPrototypeKey(key: string): boolean {
  return key === "__proto__" || key === "constructor" || key === "prototype";
}

function mergeSyncedValue(serverValue: unknown, localValue: unknown, localChange?: PreferenceChange): unknown {
  if (localChange === undefined) {
    return serverValue;
  }

  if (localChange === true || !isPlainObject(serverValue) || !isPlainObject(localValue)) {
    return localValue;
  }

  const mergedValue: Record<string, unknown> = {};
  const keys = new Set([...Object.keys(serverValue), ...Object.keys(localValue)]);

  for (const key of keys) {
    if (isPrototypeKey(key)) {
      continue;
    }

    const nestedChange = localChange[key];
    if (nestedChange !== undefined) {
      mergedValue[key] = mergeSyncedValue(serverValue[key], localValue[key], nestedChange);
      continue;
    }

    if (Object.hasOwn(serverValue, key)) {
      mergedValue[key] = serverValue[key];
    } else {
      mergedValue[key] = localValue[key];
    }
  }

  return mergedValue;
}

function mergeSyncedSettings(
  settings: Record<string, unknown>,
  preferences: LocationViewPreferences,
  localChanges: PreferenceChanges = {}
): LocationViewPreferences {
  const nextPreferences = { ...preferences };

  forEachSyncedPreference(key => {
    if (key in settings) {
      nextPreferences[key] = mergeSyncedValue(settings[key], preferences[key], localChanges[key]) as never;
    }
  });

  return nextPreferences;
}

function cloneSyncedSettings(settings: Record<string, unknown>): Record<string, unknown> {
  return JSON.parse(JSON.stringify(settings)) as Record<string, unknown>;
}

function getPreferenceChange(previousValue: unknown, nextValue: unknown): PreferenceChange | null {
  if (JSON.stringify(previousValue) === JSON.stringify(nextValue)) {
    return null;
  }

  if (isPlainObject(previousValue) && isPlainObject(nextValue)) {
    const changedFields: Record<string, PreferenceChange> = {};
    const keys = new Set([...Object.keys(previousValue), ...Object.keys(nextValue)]);

    for (const key of keys) {
      if (isPrototypeKey(key)) {
        continue;
      }

      const nestedChange = getPreferenceChange(previousValue[key], nextValue[key]);
      if (nestedChange !== null) {
        changedFields[key] = nestedChange;
      }
    }

    if (Object.keys(changedFields).length > 0) {
      return changedFields;
    }
  }

  return true;
}

function getChangedPreferences(
  previousSettings: Record<string, unknown>,
  preferences: LocationViewPreferences
): PreferenceChanges {
  const changedPreferences: PreferenceChanges = {};

  forEachSyncedPreference(key => {
    const change = getPreferenceChange(previousSettings[key], preferences[key]);
    if (change !== null) {
      changedPreferences[key] = change;
    }
  });

  return changedPreferences;
}

export function configureViewPreferenceSync(config: PreferenceSyncConfig) {
  syncConfig = {
    ...syncConfig,
    ...config,
  };
}

async function fetchViewPreferencesFromServer(): Promise<Record<string, unknown> | null> {
  const auth = useAuthContext();
  if (!auth.isAuthorized()) {
    return null;
  }

  const api = useUserApi();
  const { data, error } = await api.user.getSettings();
  if (error || !data?.item) {
    return null;
  }

  return data.item;
}
export function useViewPreferencesSync() {
  if (syncInitialized || !import.meta.client) {
    return;
  }

  syncInitialized = true;

  const auth = useAuthContext();
  const preferences = results as unknown as Ref<LocationViewPreferences>;
  let pauseServerSaves = true;
  let applyingServerSnapshot = false;
  let saveInFlight = false;
  let refreshInFlight = false;
  let refreshRequested = false;
  let localRevision = 0;
  let syncedRevision = 0;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;

  const scheduleRetry = () => {
    if (retryTimer !== null) {
      return;
    }

    retryTimer = setTimeout(() => {
      retryTimer = null;
      void saveToServer();
    }, 1000);
  };

  const markDirty = () => {
    localRevision += 1;
    queueSaveToServer();
  };

  const saveToServer = async () => {
    if (saveInFlight || retryTimer !== null || pauseServerSaves || !auth.isAuthorized()) {
      return;
    }

    saveInFlight = true;

    const api = useUserApi();
    try {
      while (syncedRevision < localRevision && !pauseServerSaves && auth.isAuthorized()) {
        const targetRevision = localRevision;
        let error = false;
        try {
          ({ error } = await api.user.setSettings(buildSyncedSettings(preferences.value)));
        } catch {
          scheduleRetry();
          return;
        }

        if (error) {
          scheduleRetry();
          return;
        }

        syncedRevision = targetRevision;
      }
    } finally {
      saveInFlight = false;

      if (syncedRevision < localRevision && retryTimer === null && !pauseServerSaves) {
        void saveToServer();
      }
    }
  };

  const queueSaveToServer = useDebounceFn(() => {
    void saveToServer();
  }, 400);

  const refreshFromServer = async () => {
    refreshRequested = true;
    if (refreshInFlight) {
      return;
    }

    refreshInFlight = true;
    try {
      while (refreshRequested) {
        refreshRequested = false;
        const refreshRevision = localRevision;
        const refreshSettings = cloneSyncedSettings(buildSyncedSettings(preferences.value));

        pauseServerSaves = true;
        try {
          const settings = await fetchViewPreferencesFromServer();
          if (settings) {
            const localChanges =
              localRevision === refreshRevision ? {} : getChangedPreferences(refreshSettings, preferences.value);
            applyingServerSnapshot = true;
            preferences.value = mergeSyncedSettings(settings, preferences.value, localChanges);
          }
        } finally {
          applyingServerSnapshot = false;
          pauseServerSaves = false;
        }

        if (syncedRevision < localRevision) {
          await saveToServer();
        }
      }
    } finally {
      refreshInFlight = false;
    }
  };

  watch(
    preferences,
    () => {
      if (applyingServerSnapshot) {
        return;
      }

      markDirty();
    },
    { deep: true, flush: "sync" }
  );

  watch(
    () => auth.token,
    token => {
      if (!token) {
        pauseServerSaves = true;
        syncedRevision = localRevision;
        return;
      }

      void refreshFromServer();
    },
    { immediate: true }
  );

  onServerEvent(ServerEvent.UserMutation, () => {
    void refreshFromServer();
  });
}

export function useViewPreferences(): Ref<LocationViewPreferences> {
  // casting is required because the type returned is removable, however since we
  // use `mergeDefaults` the result _should_ always be present.
  return results as unknown as Ref<LocationViewPreferences>;
}

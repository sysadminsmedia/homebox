import type { Ref } from "vue";
import type { ItemSummary } from "~/lib/api/types/data-contracts";
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
    value: keyof ItemSummary;
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

const DEFAULT_PREFERENCES: LocationViewPreferences = {
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

function mergeSyncedSettings(
  settings: Record<string, unknown>,
  preferences: LocationViewPreferences
): LocationViewPreferences {
  const nextPreferences = { ...preferences };

  forEachSyncedPreference(key => {
    if (key in settings) {
      nextPreferences[key] = settings[key] as never;
    }
  });

  return nextPreferences;
}

export function configureViewPreferenceSync(config: PreferenceSyncConfig) {
  syncConfig = {
    ...syncConfig,
    ...config,
  };
}

async function refreshViewPreferencesFromServer(preferences: Ref<LocationViewPreferences>) {
  const auth = useAuthContext();
  if (!auth.isAuthorized()) {
    return;
  }

  const api = useUserApi();
  const { data, error } = await api.user.getSettings();
  if (error || !data?.item) {
    return;
  }

  preferences.value = mergeSyncedSettings(data.item, preferences.value);
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
    if (saveInFlight || pauseServerSaves || !auth.isAuthorized()) {
      return;
    }

    saveInFlight = true;

    const api = useUserApi();
    try {
      while (syncedRevision < localRevision && !pauseServerSaves && auth.isAuthorized()) {
        const targetRevision = localRevision;
        const { error } = await api.user.setSettings(buildSyncedSettings(preferences.value));
        if (error) {
          scheduleRetry();
          return;
        }

        syncedRevision = targetRevision;
      }
    } finally {
      saveInFlight = false;

      if (syncedRevision < localRevision && !pauseServerSaves) {
        void saveToServer();
      }
    }
  };

  const queueSaveToServer = useDebounceFn(() => {
    void saveToServer();
  }, 400);

  const refreshFromServer = async () => {
    pauseServerSaves = true;
    applyingServerSnapshot = true;
    try {
      await refreshViewPreferencesFromServer(preferences);
    } finally {
      applyingServerSnapshot = false;
    }
    pauseServerSaves = false;

    if (syncedRevision < localRevision) {
      void saveToServer();
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
    { deep: true }
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

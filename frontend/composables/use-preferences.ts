import type { Ref } from "vue";
import {
  type LocationViewPreferences,
  type PreferenceSyncConfig,
  DEFAULT_PREFERENCES,
  buildSyncedSettings,
  mergeSyncedSettings,
} from "./preferences-utils";

export type { ViewType, DuplicateSettings, LocationViewPreferences, PreferenceSyncConfig } from "./preferences-utils";
export { DEFAULT_PREFERENCES, buildSyncedSettings, mergeSyncedSettings } from "./preferences-utils";

let syncConfig: PreferenceSyncConfig = {
  itemDisplayView: false,
  shownMultiTabWarning: false,
};

let syncInitialized = false;

const results = useLocalStorage("homebox/preferences/location", DEFAULT_PREFERENCES, { mergeDefaults: true });

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

  preferences.value = mergeSyncedSettings(data.item, preferences.value, syncConfig);
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
    if (saveInFlight || pauseServerSaves || !auth.isAuthorized()) {
      return;
    }

    saveInFlight = true;

    const api = useUserApi();
    try {
      while (syncedRevision < localRevision && !pauseServerSaves && auth.isAuthorized()) {
        const targetRevision = localRevision;
        const { data: current } = await api.user.getSettings();
        const merged = { ...(current?.item ?? {}), ...buildSyncedSettings(preferences.value, syncConfig) };
        const { error } = await api.user.setSettings(merged);
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
    refreshRequested = true;
    if (refreshInFlight) {
      return;
    }

    refreshInFlight = true;
    try {
      while (refreshRequested) {
        refreshRequested = false;

        pauseServerSaves = true;
        applyingServerSnapshot = true;
        try {
          await refreshViewPreferencesFromServer(preferences);
        } finally {
          applyingServerSnapshot = false;
        }
        pauseServerSaves = false;

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

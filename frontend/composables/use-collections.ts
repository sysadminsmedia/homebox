import { ref, computed } from "vue";
import { useUserApi } from "~/composables/use-api";
import { useViewPreferences } from "~/composables/use-preferences";
import { useRoute, navigateTo } from "#imports";

export type CollectionSummary = {
  id: string;
  name: string;
};

const collections = ref<CollectionSummary[]>([]);
const selectedId = ref<string | null>(null);
const refreshing = ref(false);

export const useCollections = () => {
  const load = async () => {
    if (window.location.pathname === "/") {
      console.debug("[useCollections] On root path '/', skipping load");
      return;
    }

    if (refreshing.value) {
      console.debug("[useCollections] Load already in progress, skipping");
      return;
    }
    refreshing.value = true;
    console.debug("[useCollections] Starting load");

    try {
      const api = useUserApi();
      const prefs = useViewPreferences();

      console.debug("[useCollections] Fetching current user");
      const { data: userResp } = await api.user.self();
      const user = userResp?.item;
      console.debug("[useCollections] User loaded", { userId: user?.id, defaultGroupId: user?.defaultGroupId });

      console.debug("[useCollections] Fetching all groups");
      const { data: allGroups } = await api.group.getAll();
      const available = Array.isArray(allGroups)
        ? (allGroups as Array<{ id: string; name: string }>).map(g => ({ id: g.id, name: g.name }))
        : [];

      collections.value = available;
      console.debug("[useCollections] Collections loaded", {
        count: collections.value.length,
        collections: collections.value,
      });

      try {
        const route = useRoute();
        if (collections.value.length === 0) {
          console.warn("[useCollections] No collections available for user");
          if (import.meta.client && route.path !== "/no-collections" && route.path !== "/") {
            console.log("[useCollections] Navigating to /no-collections (no available collections)");
            void navigateTo("/no-collections");
          }
        }
      } catch (e) {
        console.error("[useCollections] Navigation error:", e);
      }

      const prefId = prefs.value.collectionId ?? null;
      console.debug("[useCollections] Selection preference:", { prefId });

      if (prefId && collections.value.find(c => c.id === prefId)) {
        selectedId.value = prefId;
        console.debug("[useCollections] Using preferred collection", { selectedId: prefId });
      } else if (user?.defaultGroupId && collections.value.find(c => c.id === user.defaultGroupId)) {
        selectedId.value = user.defaultGroupId;
        prefs.value.collectionId = selectedId.value;
        console.debug("[useCollections] Using user default collection", { selectedId: user.defaultGroupId });
      } else if (collections.value.length > 0) {
        const first = collections.value[0];
        if (first) {
          selectedId.value = first.id;
          prefs.value.collectionId = selectedId.value;
          console.debug("[useCollections] Using first available collection", { selectedId: first.id });
        }
      } else {
        selectedId.value = null;
        prefs.value.collectionId = null;
        console.warn("[useCollections] No collection selected - empty list");
      }
    } catch (e) {
      console.error("[useCollections] Error loading collections:", e);
    } finally {
      refreshing.value = false;
      console.debug("[useCollections] Load complete", { selectedId: selectedId.value });
    }
  };

  const set = (id: string | null) => {
    selectedId.value = id;
    try {
      const prefs = useViewPreferences();
      prefs.value.collectionId = id;
    } catch (e) {
      // ignore
    }
  };

  const clear = () => {
    set(null);
  };

  const selectedCollection = computed(() => collections.value.find(c => c.id === selectedId.value) ?? null);

  return {
    collections,
    selectedId,
    selectedCollection,
    refreshing,
    load,
    set,
    clear,
  };
};

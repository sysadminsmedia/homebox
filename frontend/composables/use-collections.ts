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
    if (refreshing.value) return;
    refreshing.value = true;

    try {
      const api = useUserApi();
      const prefs = useViewPreferences();

      const { data: userResp } = await api.user.self();
      const user = userResp?.item;

      const { data: allGroups } = await api.group.get();
      const available = Array.isArray(allGroups)
        ? (allGroups as Array<{ id: string; name: string }>).map(g => ({ id: g.id, name: g.name }))
        : [];

      collections.value = available;

      try {
        const route = useRoute();
        if (collections.value.length === 0) {
          if (import.meta.client && route.path !== "/no-collections") {
            void navigateTo("/no-collections");
          }
        }
      } catch (e) {
        console.error("Navigation error in useCollections:", e);
      }
      const prefId = prefs.value.collectionId ?? null;
      if (prefId && collections.value.find(c => c.id === prefId)) {
        selectedId.value = prefId;
      } else if (user?.defaultGroupId && collections.value.find(c => c.id === user.defaultGroupId)) {
        selectedId.value = user.defaultGroupId;
        prefs.value.collectionId = selectedId.value;
      } else if (collections.value.length > 0) {
        const first = collections.value[0];
        if (first) {
          selectedId.value = first.id;
          prefs.value.collectionId = selectedId.value;
        }
      } else {
        selectedId.value = null;
        prefs.value.collectionId = null;
      }
    } finally {
      refreshing.value = false;
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

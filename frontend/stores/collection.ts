import { defineStore } from "pinia";
import { useViewPreferences } from "~/composables/use-preferences";

/** Minimal collection summary used by UI */
export type CollectionSummary = {
  id: string;
  name: string;
};

export const useCollectionStore = defineStore("collection", {
  state: () => ({
    id: null as string | null,
    collections: [] as CollectionSummary[],
    refreshing: false,
  }),
  getters: {
    selectedCollection(state): CollectionSummary | null {
      return state.collections.find(c => c.id === state.id) ?? null;
    },
  },
  actions: {
    async load() {
      if (this.refreshing) return;
      this.refreshing = true;

      const api = useUserApi();
      const prefs = useViewPreferences();

      try {
        const { data: userResp } = await api.user.self();
        if (!userResp?.item) return;

        const user = userResp.item;

        const { data: allGroups } = await api.group.get();

        const available = Array.isArray(allGroups)
          ? (allGroups as Array<{ id: string; name: string }>).map(g => ({ id: g.id, name: g.name }))
          : [];

        this.collections = available;

        // Determine selected collection from preferences (if still present in list),
        // otherwise use user's defaultGroupId if available and present, otherwise first available.
        const prefId = prefs.value.collectionId ?? null;
        if (prefId && this.collections.find(c => c.id === prefId)) {
          this.id = prefId;
        } else if (user.defaultGroupId && this.collections.find(c => c.id === user.defaultGroupId)) {
          this.id = user.defaultGroupId;
          prefs.value.collectionId = this.id;
        } else if (this.collections.length > 0) {
          const first = this.collections[0];
          if (first) {
            this.id = first.id;
            prefs.value.collectionId = this.id;
          }
        } else {
          this.id = null;
          prefs.value.collectionId = null;
        }
      } finally {
        this.refreshing = false;
      }
    },

    set(id: string | null) {
      this.id = id;
      try {
        const prefs = useViewPreferences();
        prefs.value.collectionId = id;
      } catch (e) {
        // ignore
      }
    },

    clear() {
      this.set(null);
    },
  },
});

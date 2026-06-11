import { defineStore } from "pinia";
import type { EntityTypeSummary } from "~~/lib/api/types/data-contracts";

export const useEntityTypeStore = defineStore("entityTypes", {
  state: () => ({
    types: null as EntityTypeSummary[] | null,
    client: useUserApi(),
    refreshPromise: null as Promise<void> | null,
  }),
  getters: {
    allTypes(state): EntityTypeSummary[] {
      return state.types ?? [];
    },
    locationTypes(state): EntityTypeSummary[] {
      return (state.types ?? []).filter(t => t.isLocation);
    },
    itemTypes(state): EntityTypeSummary[] {
      return (state.types ?? []).filter(t => !t.isLocation);
    },
  },
  actions: {
    async ensureFetched() {
      if (this.types !== null) return;
      if (this.refreshPromise === null) {
        this.refreshPromise = this.refresh().then(() => {});
      }
      await this.refreshPromise;
    },

    async refresh() {
      const result = await this.client.entityTypes.getAll();
      if (result.error) {
        return result;
      }
      this.types = result.data ?? [];
      return result;
    },

    findById(id: string) {
      return (this.types ?? []).find(t => t.id === id) ?? null;
    },
  },
});

export default useEntityTypeStore;

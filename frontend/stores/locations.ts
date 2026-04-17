import { defineStore } from "pinia";
import type { ItemsApi } from "~~/lib/api/classes/items";
import type { EntitySummary, TreeItem } from "~~/lib/api/types/data-contracts";

export const useLocationStore = defineStore("locations", {
  state: () => ({
    parents: null as EntitySummary[] | null,
    Locations: null as EntitySummary[] | null,
    client: useUserApi(),
    tree: null as TreeItem[] | null,
    refreshLocationsPromise: null as Promise<void> | null,
  }),
  getters: {
    /**
     * locations represents the locations that are currently in the store. The store is
     * synched with the server by intercepting the API calls and updating on the
     * response
     */
    parentLocations(state): EntitySummary[] {
      if (state.parents === null) {
        this.client.items.getLocations({ filterChildren: true }).then(result => {
          if (result.error) {
            console.error(result.error);
            return;
          }

          this.parents = result.data;
        });
      }
      return state.parents ?? [];
    },
    allLocations(state): EntitySummary[] {
      return state.Locations ?? [];
    },
  },
  actions: {
    async ensureLocationsFetched() {
      if (this.Locations !== null) {
        return;
      }

      if (this.refreshLocationsPromise === null) {
        this.refreshLocationsPromise = this.refreshChildren().then(() => {});
      }
      await this.refreshLocationsPromise;
    },
    async refreshParents(): ReturnType<ItemsApi["getLocations"]> {
      const result = await this.client.items.getLocations({ filterChildren: true });
      if (result.error) {
        return result;
      }

      this.parents = result.data;
      return result;
    },
    async refreshChildren(): ReturnType<ItemsApi["getLocations"]> {
      const result = await this.client.items.getLocations({ filterChildren: false });
      if (result.error) {
        return result;
      }

      this.Locations = result.data;
      return result;
    },
    async refreshTree(): ReturnType<ItemsApi["getTree"]> {
      const result = await this.client.items.getTree();
      if (result.error) {
        return result;
      }

      this.tree = result.data;
      return result;
    },
  },
});

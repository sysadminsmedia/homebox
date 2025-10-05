import { defineStore } from "pinia";
import type { LocationsApi } from "~~/lib/api/classes/locations";
import type { LocationOutCount, TreeItem } from "~~/lib/api/types/data-contracts";

export const useLocationStore = defineStore("locations", {
  state: () => ({
    parents: null as LocationOutCount[] | null,
    Locations: null as LocationOutCount[] | null,
    client: useUserApi(),
    tree: null as TreeItem[] | null,
  }),
  getters: {
    /**
     * locations represents the locations that are currently in the store. The store is
     * synched with the server by intercepting the API calls and updating on the
     * response
     */
    parentLocations(state): LocationOutCount[] {
      if (state.parents === null) {
        this.client.locations.getAll({ filterChildren: true }).then(result => {
          if (result.error) {
            console.error(result.error);
            return;
          }

          this.parents = result.data;
        });
      }
      return state.parents ?? [];
    },
    allLocations(state): LocationOutCount[] {
      // ensures that locations are eventually available but not synchronously
      state.ensureLocationsFetched()
      return state.Locations ?? [];
    },
  },
  actions: {
    async ensureLocationsFetched() {
      if (this.Locations !== null) {
          return;
      }

      if (this.refreshLocationsPromise === undefined) {
          this.refreshLocationsPromise = this.refreshChildren()
      }
      await this.refreshLocationsPromise;
    },
    async refreshParents(): ReturnType<LocationsApi["getAll"]> {
      const result = await this.client.locations.getAll({ filterChildren: true });
      if (result.error) {
        return result;
      }

      this.parents = result.data;
      return result;
    },
    async refreshChildren(): ReturnType<LocationsApi["getAll"]> {
      const result = await this.client.locations.getAll({ filterChildren: false });
      if (result.error) {
        return result;
      }

      this.Locations = result.data;
      return result;
    },
    async refreshTree(): ReturnType<LocationsApi["getTree"]> {
      const result = await this.client.locations.getTree();
      if (result.error) {
        return result;
      }

      this.tree = result.data;
      return result;
    },
  },
});

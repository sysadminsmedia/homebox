import { defineStore } from "pinia";
import type { TagOut } from "~~/lib/api/types/data-contracts";

export const useTagStore = defineStore("tags", {
  state: () => ({
    allTags: null as TagOut[] | null,
    client: useUserApi(),
    refreshAllTagsPromise: null as Promise<void> | null,
  }),
  getters: {
    /**
     * tags represents the tags that are currently in the store. The store is
     * synched with the server by intercepting the API calls and updating on the
     * response.
     */
    tags(state): TagOut[] {
      return state.allTags ?? [];
    },
  },
  actions: {
    async ensureAllTagsFetched() {
      if (this.allTags !== null) {
        return;
      }

      if (this.refreshAllTagsPromise === null) {
        this.refreshAllTagsPromise = this.refresh().then(() => {});
      }
      await this.refreshAllTagsPromise;
    },
    async refresh() {
      const result = await this.client.tags.getAll();
      if (result.error) {
        return result;
      }

      this.allTags = result.data;
      return result;
    },
  },
});

import { defineStore } from "pinia";
import type { LabelOut } from "~~/lib/api/types/data-contracts";

export const useLabelStore = defineStore("labels", {
  state: () => ({
    allLabels: null as LabelOut[] | null,
    client: useUserApi(),
    refreshAllLabelsPromise: null as Promise<void> | null,
  }),
  getters: {
    /**
     * labels represents the labels that are currently in the store. The store is
     * synched with the server by intercepting the API calls and updating on the
     * response.
     */
    labels(state): LabelOut[] {
      return state.allLabels ?? [];
    },
  },
  actions: {
    async ensureAllLabelsFetched() {
      if (this.allLabels !== null) {
        return;
      }

      if (this.refreshAllLabelsPromise === null) {
        this.refreshAllLabelsPromise = this.refresh().then(() => {});
      }
      await this.refreshAllLabelsPromise;
    },
    async refresh() {
      const result = await this.client.labels.getAll();
      if (result.error) {
        return result;
      }

      this.allLabels = result.data;
      return result;
    },
  },
});

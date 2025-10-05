import { defineStore } from "pinia";
import type { LabelOut } from "~~/lib/api/types/data-contracts";

export const useLabelStore = defineStore("labels", {
  state: () => ({
    allLabels: null as LabelOut[] | null,
    client: useUserApi(),
  }),
  getters: {
    /**
     * labels represents the labels that are currently in the store. The store is
     * synched with the server by intercepting the API calls and updating on the
     * response.
     */
    labels(state): LabelOut[] {
      // ensures that labels are eventually available but not synchronously
      state.ensureAllLabelsFetched()
      return state.allLabels ?? [];
    },
  },
  actions: {
    async ensureAllLabelsFetched() {
      if (this.allLabels !== null) {
          return;
      }

      if (this.refreshAllLabelsPromise === undefined) {
          this.refreshAllLabelsPromise = this.refresh()
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

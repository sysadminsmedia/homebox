import type { ItemSummary, LabelSummary, LocationSummary } from "~~/lib/api/types/data-contracts";
import type { UserClient } from "~~/lib/api/user";

type SearchOptions = {
  immediate?: boolean;
};

export function useItemSearch(client: UserClient, opts?: SearchOptions) {
  const query = ref("");
  const locations = ref<LocationSummary[]>([]);
  const labels = ref<LabelSummary[]>([]);
  const results = ref<ItemSummary[]>([]);
  const includeArchived = ref(false);
  const isLoading = ref(false);
  const pendingQuery = ref<string | null>(null);

  watchDebounced(query, search, { debounce: 250, maxWait: 1000 });
  async function search(): Promise<boolean> {
    if (isLoading.value) {
      // Store the latest query to run after current search completes
      pendingQuery.value = query.value;
      return false;
    }

    const searchQuery = query.value;
    isLoading.value = true;
    try {
      const locIds = locations.value.map(l => l.id);
      const labelIds = labels.value.map(l => l.id);

      const { data, error } = await client.items.getAll({
        q: searchQuery,
        locations: locIds,
        labels: labelIds,
        includeArchived: includeArchived.value,
      });

      if (error || !data) {
        console.error("useItemSearch.search error:", error);
        return false;
      }

      results.value = data.items ?? [];
      return true;
    } finally {
      isLoading.value = false;

      // If user changed query while we were searching, run again with the latest query
      if (pendingQuery.value !== null && pendingQuery.value !== searchQuery) {
        const nextQuery = pendingQuery.value;
        pendingQuery.value = null;
        // Use nextTick to avoid potential recursion issues
        await nextTick();
        if (query.value === nextQuery) {
          await search();
        }
      } else {
        pendingQuery.value = null;
      }
    }
  }

  async function triggerSearch(): Promise<boolean> {
    try {
      return await search();
    } catch (err) {
      console.error("triggerSearch error:", err);
      return false;
    }
  }

  if (opts?.immediate) {
    search()
      .then(success => {
        if (!success) {
          console.error("Initial search failed");
        }
      })
      .catch(err => {
        console.error("Initial search error:", err);
      });
  }

  return {
    query,
    results,
    locations,
    labels,
    isLoading,
    triggerSearch,
  };
}

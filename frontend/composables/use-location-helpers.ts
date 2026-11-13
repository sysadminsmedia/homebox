import fuzzysort from "fuzzysort";
import type { Ref } from "vue";
import type { EntitySummary, TreeItem } from "~~/lib/api/types/data-contracts";

export interface FlatTreeItem {
  assetId: string;
  id: string;
  name: string;
  treeString: string;
}

function flatTree(tree: TreeItem[]): FlatTreeItem[] {
  const v = [] as FlatTreeItem[];

  // turns the nested items into a flat items array where
  // the display is a string of the tree hierarchy separated by breadcrumbs

  function flatten(items: TreeItem[], display: string) {
    if (!items) {
      return;
    }

    for (const item of items) {
      v.push({
        assetId: "",
        id: item.id,
        name: item.name,
        treeString: display + item.name,
      });
      if (item.children) {
        flatten(item.children, display + item.name + " > ");
      }
    }
  }

  flatten(tree, "");

  return v;
}

export function withAssetIds(
  locations: Omit<FlatTreeItem, "assetId">[],
  summaries: Pick<EntitySummary, "assetId" | "id">[]
): FlatTreeItem[] {
  const assetIdsByLocationId = new Map(summaries.map(summary => [summary.id, summary.assetId]));

  return locations.map(location => ({
    ...location,
    assetId: assetIdsByLocationId.get(location.id) ?? "",
  }));
}

export function filterLocations(search: string, locations: FlatTreeItem[]): FlatTreeItem[] {
  const isAssetIdSearch = search.startsWith("#");
  const query = isAssetIdSearch ? search.slice(1) : search;
  const keys = isAssetIdSearch ? ["assetId"] : ["name", "treeString"];

  return fuzzysort.go(query, locations, { keys, all: true }).map(result => result.obj);
}

function filterOutSubtree(tree: TreeItem[], excludeId: string): TreeItem[] {
  // Recursively filters out a subtree starting from excludeId
  const result: TreeItem[] = [];

  for (const item of tree) {
    if (item.id === excludeId) {
      continue;
    }

    const newItem = { ...item };
    if (item.children) {
      newItem.children = filterOutSubtree(item.children, excludeId);
    }

    result.push(newItem);
  }

  return result;
}

export function useFlatLocations(excludeSubtreeForLocation?: EntitySummary): Ref<FlatTreeItem[]> {
  const locations = useLocationStore();

  if (locations.tree === null) {
    locations.refreshTree();
  }
  void locations.ensureLocationsFetched();

  return computed(() => {
    if (locations.tree === null) {
      return [];
    }

    const filteredTree = excludeSubtreeForLocation
      ? filterOutSubtree(locations.tree, excludeSubtreeForLocation.id)
      : locations.tree;

    return withAssetIds(flatTree(filteredTree), locations.allLocations);
  });
}

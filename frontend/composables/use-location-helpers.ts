import type { Ref } from "vue";
import type { LocationSummary, TreeItem } from "~~/lib/api/types/data-contracts";

export interface FlatTreeItem {
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

export function useFlatLocations(excludeSubtreeForLocation?: LocationSummary): Ref<FlatTreeItem[]> {
  const locations = useLocationStore();

  if (locations.tree === null) {
    locations.refreshTree();
  }

  return computed(() => {
    if (locations.tree === null) {
      return [];
    }

    const filteredTree = excludeSubtreeForLocation
      ? filterOutSubtree(locations.tree, excludeSubtreeForLocation.id)
      : locations.tree;

    return flatTree(filteredTree);
  });
}

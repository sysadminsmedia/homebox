import { ref, type Ref } from "vue";

type TreeState = Record<string, boolean>;

const store: Record<string, Ref<TreeState>> = {};

export function newTreeKey(): string {
  return Math.random().toString(36).substring(2);
}

export function useTreeState(key: string): Ref<TreeState> {
  if (!store[key]) {
    store[key] = ref({});
  }

  return store[key];
}

const showItemsStore: Record<string, Ref<boolean>> = {};

/**
 * Whether a location tree lists items as well as locations.
 *
 * Kept in the same module-level store as the tree's expanded state, so it
 * lasts exactly as long: leaving the Locations page and coming back keeps
 * both. It used to be a ref local to the page, so Expand/Collapse survived a
 * visit to another page while Show/Hide Items reset to "show" every time
 * (#1577).
 */
export function useShowItems(key: string): Ref<boolean> {
  if (!showItemsStore[key]) {
    showItemsStore[key] = ref(true);
  }

  return showItemsStore[key];
}

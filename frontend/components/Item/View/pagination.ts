import type { ShallowUnwrapRef } from "vue";

export type Pagination = ShallowUnwrapRef<{
  page: WritableComputedRef<number, number>;
  pageSize: ComputedRef<number>;
  totalSize: Ref<number, number>;
  setPage: (newPage: number) => void;
}>;

// Keep a page selection within the current result set after data mutations.
export function clampPageToTotal(page: number, total: number, pageSize: number): number {
  const lastPage = Math.max(1, Math.ceil(total / Math.max(1, pageSize)));
  return Math.min(Math.max(1, page), lastPage);
}

// Cached async-data from a prior page must never be shown under a new page
// number, including during refreshes or quick navigation.
export function isCurrentLocationPage(status: string, requestedKey: string, loadedKey?: string): boolean {
  return status === "success" && requestedKey === loadedKey;
}

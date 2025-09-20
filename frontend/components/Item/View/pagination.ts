import type { ShallowUnwrapRef } from "vue";

export type Pagination = ShallowUnwrapRef<{
  page: globalThis.WritableComputedRef<number, number>;
  pageSize: globalThis.ComputedRef<number>;
  totalSize: globalThis.Ref<number, number>;
  setPage: (newPage: number) => void;
}>;

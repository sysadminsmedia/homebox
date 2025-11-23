import type { ShallowUnwrapRef } from "vue";

export type Pagination = ShallowUnwrapRef<{
  page: WritableComputedRef<number, number>;
  pageSize: ComputedRef<number>;
  totalSize: Ref<number, number>;
  setPage: (newPage: number) => void;
}>;

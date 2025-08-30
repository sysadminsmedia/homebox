import { useRouteQuery as useRouteQueryBase } from "@vueuse/router";

/* eslint no-redeclare: 0 */
import type { WritableComputedRef } from "vue";

export function useRouteQuery(q: string, def: string[]): WritableComputedRef<string[]>;
export function useRouteQuery(q: string, def: string): WritableComputedRef<string>;
export function useRouteQuery(q: string, def: boolean): WritableComputedRef<boolean>;
export function useRouteQuery(q: string, def: number): WritableComputedRef<number>;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useRouteQuery(q: string, def: any): WritableComputedRef<any> {
  const route = useRoute();
  const router = useRouter();

  const v = useRouteQueryBase(q, def);

  const first = computed<string>(() => {
    const qv = route.query[q];
    if (Array.isArray(qv)) {
      return qv[0]?.toString() || def;
    }
    return qv?.toString() || def;
  });

  onMounted(() => {
    if (route.query[q] === undefined) {
      v.value = def;
    }
  });

  switch (typeof def) {
    case "string":
      return computed({
        get: () => {
          const qv = first.value;
          if (Array.isArray(qv)) {
            return qv[0];
          }
          return qv;
        },
        set: v => {
          const query = { ...route.query, [q]: v };
          router.push({ query });
        },
      });
    case "object": // array
      return computed({
        get: () => {
          const qv = route.query[q];
          if (Array.isArray(qv)) {
            return qv;
          }
          return [qv];
        },
        set: v => {
          const query = { ...route.query, [q]: v };
          router.push({ query });
        },
      });
    case "boolean":
      return computed({
        get: () => {
          return first.value === "true";
        },
        set: v => {
          const query = { ...route.query, [q]: `${v}` };
          router.push({ query });
        },
      });
    case "number":
      return computed({
        get: () => parseInt(first.value, 10),
        set: nv => {
          v.value = nv.toString();
        },
      });
  }

  throw new Error("Invalid type");
}

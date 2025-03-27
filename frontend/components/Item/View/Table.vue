<template>
  <BaseCard>
    <table class="table w-full">
      <thead>
        <tr>
          <th
            v-for="h in headers.filter(h => h.enabled)"
            :key="h.value"
            class="text-no-transform cursor-pointer bg-neutral text-sm text-neutral-content"
            @click="sortBy(h.value)"
          >
            <div
              class="flex items-center gap-1"
              :class="{
                'justify-center': h.align === 'center',
                'justify-start': h.align === 'right',
                'justify-end': h.align === 'left',
              }"
            >
              <template v-if="typeof h === 'string'">{{ h }}</template>
              <template v-else>{{ $t(h.text) }}</template>
              <div
                v-if="sortByProperty === h.value"
                :class="`inline-flex ${sortByProperty === h.value ? '' : 'opacity-0'}`"
              >
                <span class="swap swap-rotate" :class="{ 'swap-active': pagination.descending }">
                  <MdiArrowDown class="swap-on size-5" />
                  <MdiArrowUp class="swap-off size-5" />
                </span>
              </div>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(d, i) in data" :key="d.id" class="hover cursor-pointer" @click="navigateTo(`/item/${d.id}`)">
          <td
            v-for="h in headers.filter(h => h.enabled)"
            :key="`${h.value}-${i}`"
            class="bg-base-100"
            :class="{
              'text-center': h.align === 'center',
              'text-right': h.align === 'right',
              'text-left': h.align === 'left',
            }"
          >
            <template v-if="h.type === 'name'">
              <NuxtLink class="hover text-wrap" :to="`/item/${d.id}`">
                {{ d.name }}
              </NuxtLink>
            </template>
            <template v-else-if="h.type === 'price'">
              <Currency :amount="d.purchasePrice" />
            </template>
            <template v-else-if="h.type === 'boolean'">
              <MdiCheck v-if="d.insured" class="inline size-5 text-green-500" />
              <MdiClose v-else class="inline size-5 text-red-500" />
            </template>
            <template v-else-if="h.type === 'location'">
              <NuxtLink v-if="d.location" class="hover:link" :to="`/location/${d.location.id}`">
                {{ d.location.name }}
              </NuxtLink>
            </template>
            <template v-else-if="h.type === 'date'">
              <DateTime :date="d[h.value]" datetime-type="date" />
            </template>
            <slot v-else :name="cell(h)" v-bind="{ item: d }">
              {{ extractValue(d, h.value) }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
    <div
      class="flex items-center justify-end gap-3 border-t p-3"
      :class="{
        hidden: disableControls,
      }"
    >
      <div class="dropdown dropdown-top dropdown-hover">
        <label tabindex="0" class="btn btn-square btn-outline btn-sm m-1">
          <MdiTableCog />
        </label>
        <ul tabindex="0" class="dropdown-content rounded-box flex w-64 flex-col gap-2 bg-base-100 p-2 pl-3 shadow">
          <li>Headers:</li>
          <li v-for="(h, i) in headers" :key="h.value" class="flex flex-row items-center gap-1">
            <button
              class="btn btn-square btn-ghost btn-xs"
              :class="{
                'btn-disabled': i === 0,
              }"
              @click="moveHeader(i, i - 1)"
            >
              <MdiArrowUp />
            </button>
            <button
              class="btn btn-square btn-ghost btn-xs"
              :class="{
                'btn-disabled': i === headers.length - 1,
              }"
              @click="moveHeader(i, i + 1)"
            >
              <MdiArrowDown />
            </button>
            <input
              :id="h.value"
              type="checkbox"
              class="checkbox checkbox-primary"
              :checked="h.enabled"
              @change="toggleHeader(h.value)"
            />
            <label class="label-text" :for="h.value"> {{ $t(h.text) }} </label>
          </li>
        </ul>
      </div>
      <div class="hidden md:block">{{ $t("components.item.view.table.rows_per_page") }}</div>
      <select v-model.number="pagination.rowsPerPage" class="select select-primary select-sm">
        <option :value="10">10</option>
        <option :value="25">25</option>
        <option :value="50">50</option>
        <option :value="100">100</option>
      </select>
      <div class="btn-group">
        <button :disabled="!hasPrev" class="btn btn-sm" @click="prev()">«</button>
        <button class="btn btn-sm">{{ $t("components.item.view.table.page") }} {{ pagination.page }}</button>
        <button :disabled="!hasNext" class="btn btn-sm" @click="next()">»</button>
      </div>
    </div>
  </BaseCard>
</template>

<script setup lang="ts">
  import type { TableData, TableHeader } from "./Table.types";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiCheck from "~icons/mdi/check";
  import MdiClose from "~icons/mdi/close";
  import MdiTableCog from "~icons/mdi/table-cog";

  type Props = {
    items: ItemSummary[];
    disableControls?: boolean;
  };
  const props = defineProps<Props>();

  const sortByProperty = ref<keyof ItemSummary | "">("");

  const preferences = useViewPreferences();

  const defaultHeaders = [
    {
      text: "items.name",
      value: "name",
      enabled: true,
      type: "name",
    },
    { text: "items.quantity", value: "quantity", align: "center", enabled: true },
    { text: "items.insured", value: "insured", align: "center", enabled: true, type: "boolean" },
    { text: "items.purchase_price", value: "purchasePrice", align: "center", enabled: true, type: "price" },
    { text: "items.location", value: "location", align: "center", enabled: false, type: "location" },
    { text: "items.archived", value: "archived", align: "center", enabled: false, type: "boolean" },
    { text: "items.created_at", value: "createdAt", align: "center", enabled: false, type: "date" },
    { text: "items.updated_at", value: "updatedAt", align: "center", enabled: false, type: "date" },
  ] satisfies TableHeader[];

  const headers = ref<TableHeader[]>(
    (preferences.value.tableHeaders ?? [])
      .concat(defaultHeaders.filter(h => !preferences.value.tableHeaders?.find(h2 => h2.value === h.value)))
      // this is a hack to make sure that any changes to the defaultHeaders are reflected in the preferences
      .map(h => ({
        ...(defaultHeaders.find(h2 => h2.value === h.value) as TableHeader),
        enabled: h.enabled,
      }))
  );

  console.log(headers.value);

  const toggleHeader = (value: string) => {
    const header = headers.value.find(h => h.value === value);
    if (header) {
      header.enabled = !header.enabled; // Toggle the 'enabled' state
    }

    preferences.value.tableHeaders = headers.value;
  };
  const moveHeader = (from: number, to: number) => {
    const header = headers.value[from];
    headers.value.splice(from, 1);
    headers.value.splice(to, 0, header);

    preferences.value.tableHeaders = headers.value;
  };

  const pagination = reactive({
    descending: false,
    page: 1,
    rowsPerPage: preferences.value.itemsPerTablePage,
    rowsNumber: 0,
  });

  watch(
    () => pagination.rowsPerPage,
    newRowsPerPage => {
      preferences.value.itemsPerTablePage = newRowsPerPage;
    }
  );

  const next = () => pagination.page++;
  const hasNext = computed<boolean>(() => {
    return pagination.page * pagination.rowsPerPage < props.items.length;
  });

  const prev = () => pagination.page--;
  const hasPrev = computed<boolean>(() => {
    return pagination.page > 1;
  });

  function sortBy(property: keyof ItemSummary) {
    if (sortByProperty.value === property) {
      pagination.descending = !pagination.descending;
    } else {
      pagination.descending = false;
    }
    sortByProperty.value = property;
  }

  function extractSortable(item: ItemSummary, property: keyof ItemSummary): string | number | boolean {
    const value = item[property];
    if (typeof value === "string") {
      // Try parse float
      const parsed = parseFloat(value);
      if (!isNaN(parsed)) {
        return parsed;
      }

      return value.toLowerCase();
    }

    if (typeof value !== "number" && typeof value !== "boolean") {
      return "";
    }

    return value;
  }

  function itemSort(a: ItemSummary, b: ItemSummary) {
    if (!sortByProperty.value) {
      return 0;
    }

    const aVal = extractSortable(a, sortByProperty.value);
    const bVal = extractSortable(b, sortByProperty.value);

    if (typeof aVal === "string" && typeof bVal === "string") {
      return aVal.localeCompare(bVal, undefined, { numeric: true, sensitivity: "base" });
    }

    if (aVal < bVal) {
      return -1;
    }
    if (aVal > bVal) {
      return 1;
    }
    return 0;
  }

  const data = computed<TableData[]>(() => {
    // sort by property
    let data = [...props.items].sort(itemSort);

    // sort descending
    if (pagination.descending) {
      data.reverse();
    }

    // paginate
    const start = (pagination.page - 1) * pagination.rowsPerPage;
    const end = start + pagination.rowsPerPage;
    data = data.slice(start, end);
    return data;
  });

  function extractValue(data: TableData, value: string) {
    const parts = value.split(".");
    let current = data;
    for (const part of parts) {
      current = current[part];
    }
    return current;
  }

  function cell(h: TableHeader) {
    return `cell-${h.value.replace(".", "_")}`;
  }
</script>

<style scoped>
  :where(.table *:first-child) :where(*:first-child) :where(th, td):first-child {
    border-top-left-radius: 0.5rem;
  }

  :where(.table *:first-child) :where(*:first-child) :where(th, td):last-child {
    border-top-right-radius: 0.5rem;
  }

  :where(.table *:last-child) :where(*:last-child) :where(th, td):first-child {
    border-bottom-left-radius: 0.5rem;
  }

  :where(.table *:last-child) :where(*:last-child) :where(th, td):last-child {
    border-bottom-right-radius: 0.5rem;
  }
</style>

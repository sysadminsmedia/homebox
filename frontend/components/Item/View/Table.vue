<template>
  <BaseCard>
    <table class="table w-full">
      <thead>
        <tr>
          <th
            v-for="h in headers"
            :key="h.value"
            class="text-no-transform text-sm bg-neutral text-neutral-content cursor-pointer"
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
              <template v-else>{{ h.text }}</template>
              <div
                v-if="sortByProperty === h.value"
                :class="`inline-flex ${sortByProperty === h.value ? '' : 'opacity-0'}`"
              >
                <span class="swap swap-rotate" :class="{ 'swap-active': pagination.descending }">
                  <MdiArrowDown class="swap-on h-5 w-5" />
                  <MdiArrowUp class="swap-off h-5 w-5" />
                </span>
              </div>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(d, i) in data" :key="d.id" class="hover cursor-pointer" @click="navigateTo(`/item/${d.id}`)">
          <td
            v-for="h in headers"
            :key="`${h.value}-${i}`"
            class="bg-base-100"
            :class="{
              'text-center': h.align === 'center',
              'text-right': h.align === 'right',
              'text-left': h.align === 'left',
            }"
          >
            <template v-if="cell(h) === 'cell-name'">
              <NuxtLink class="hover" :to="`/item/${d.id}`">
                {{ d.name }}
              </NuxtLink>
            </template>
            <template v-else-if="cell(h) === 'cell-purchasePrice'">
              <Currency :amount="d.purchasePrice" />
            </template>
            <template v-else-if="cell(h) === 'cell-insured'">
              <MdiCheck v-if="d.insured" class="text-green-500 h-5 w-5 inline" />
              <MdiClose v-else class="text-red-500 h-5 w-5 inline" />
            </template>
            <slot v-else :name="cell(h)" v-bind="{ item: d }">
              {{ extractValue(d, h.value) }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="items.length > 10" class="border-t p-3 justify-end flex gap-3">
      <div class="flex items-center">Rows per page</div>
      <select v-model.number="pagination.rowsPerPage" class="select select-sm select-primary">
        <option :value="10">10</option>
        <option :value="25">25</option>
        <option :value="50">50</option>
        <option :value="100">100</option>
      </select>
      <div class="btn-group">
        <button :disabled="!hasPrev" class="btn btn-sm" @click="prev()">«</button>
        <button class="btn btn-sm">Page {{ pagination.page }}</button>
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

  type Props = {
    items: ItemSummary[];
  };
  const props = defineProps<Props>();

  const sortByProperty = ref<keyof ItemSummary | "">("");

  const headers = computed<TableHeader[]>(() => {
    return [
      { text: "Name", value: "name" },
      { text: "Quantity", value: "quantity", align: "center" },
      { text: "Insured", value: "insured", align: "center" },
      { text: "Price", value: "purchasePrice" },
    ] as TableHeader[];
  });

  const preferences = useViewPreferences();

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

    const aLower = extractSortable(a, sortByProperty.value);
    const bLower = extractSortable(b, sortByProperty.value);

    if (aLower < bLower) {
      return -1;
    }
    if (aLower > bLower) {
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

<style scoped></style>

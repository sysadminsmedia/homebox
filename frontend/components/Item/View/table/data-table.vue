<script setup lang="ts" generic="TData, TValue">
  import BaseCard from "@/components/Base/Card.vue";
  import type { ColumnDef, SortingState, VisibilityState, ExpandedState } from "@tanstack/vue-table";
  import {
    getCoreRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    getExpandedRowModel,
    useVueTable,
  } from "@tanstack/vue-table";

  import { camelToSnakeCase, valueUpdater } from "@/lib/utils";

  import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import Button from "~/components/ui/button/Button.vue";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import Checkbox from "~/components/ui/checkbox/Checkbox.vue";
  import Label from "~/components/ui/label/Label.vue";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";

  import TableView from "./table-view.vue";
  import CardView from "./card-view.vue";
  import DataTableControls from "./data-table-controls.vue";

  const props = defineProps<{
    columns: ColumnDef<ItemSummary, TValue>[];
    data: ItemSummary[];
    disableControls?: boolean;
    view: "table" | "card";
  }>();

  const preferences = useViewPreferences();
  const defaultPageSize = preferences.value.itemsPerTablePage;
  const tableHeadersData = preferences.value.tableHeaders;
  const defaultVisible = ["name", "quantity", "insured", "purchasePrice"];

  const tableHeaders = computed(
    () =>
      tableHeadersData ??
      props.columns
        .filter(c => c.enableHiding !== false)
        .map(c => ({
          value: c.id!,
          enabled: defaultVisible.includes(c.id ?? ""),
        }))
  );

  const sorting = ref<SortingState>([]);
  const columnOrder = ref<string[]>([
    "select",
    ...(tableHeaders.value ? tableHeaders.value.map(h => h.value) : []),
    "actions",
  ]);
  const columnVisibility = ref<VisibilityState>(
    tableHeaders.value?.reduce((acc, h) => ({ ...acc, [h.value]: h.enabled }), {})
  );
  const rowSelection = ref({});
  const expanded = ref<ExpandedState>({});
  const pagination = ref({
    pageIndex: 0,
    pageSize: defaultPageSize || 12,
  });

  watch(
    () => pagination.value.pageSize,
    newSize => {
      preferences.value.itemsPerTablePage = newSize;
    }
  );

  const table = useVueTable<ItemSummary>({
    get data() {
      return props.data;
    },
    get columns() {
      return props.columns;
    },

    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getExpandedRowModel: getExpandedRowModel(),

    onSortingChange: updaterOrValue => valueUpdater(updaterOrValue, sorting),
    onColumnVisibilityChange: updaterOrValue => valueUpdater(updaterOrValue, columnVisibility),
    onRowSelectionChange: updaterOrValue => valueUpdater(updaterOrValue, rowSelection),
    onExpandedChange: updaterOrValue => valueUpdater(updaterOrValue, expanded),
    onColumnOrderChange: updaterOrValue => valueUpdater(updaterOrValue, columnOrder),
    onPaginationChange: updaterOrValue => valueUpdater(updaterOrValue, pagination),

    state: {
      get sorting() {
        return sorting.value;
      },
      get columnVisibility() {
        return columnVisibility.value;
      },
      get rowSelection() {
        return rowSelection.value;
      },
      get expanded() {
        return expanded.value;
      },
      get columnOrder() {
        return columnOrder.value;
      },
      get pagination() {
        return pagination.value;
      },
    },
  });

  const persistHeaders = () => {
    const headers = table
      .getAllColumns()
      .filter(column => column.getCanHide())
      .map(h => ({
        value: h.id as keyof ItemSummary,
        enabled: h.getIsVisible(),
      }));

    preferences.value.tableHeaders = headers;
  };

  const moveHeader = (from: number, to: number) => {
    // Only allow moving between the first and last index (excluding 'select' and 'actions')
    const start = 1; // index of 'select'
    const end = columnOrder.value.length - 2; // index before 'actions'

    if (from < start || from > end || to < start || to > end || from === to) return;

    const order = [...columnOrder.value];
    const [moved] = order.splice(from, 1);
    order.splice(to, 0, moved!);
    columnOrder.value = order;

    persistHeaders();
  };

  const toggleHeader = (id: string) => {
    const header = table
      .getAllColumns()
      .filter(column => column.getCanHide())
      .find(h => h.id === id);
    if (header) {
      header.toggleVisibility();
    }

    persistHeaders();
  };
</script>

<template>
  <div>
    <Dialog :dialog-id="DialogID.ItemTableSettings">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("components.item.view.table.table_settings") }}</DialogTitle>
        </DialogHeader>

        <div v-if="props.view === 'table'">{{ $t("components.item.view.table.headers") }}</div>
        <div v-if="props.view === 'table'" class="flex flex-col">
          <div
            v-for="(colId, i) in columnOrder.slice(1, columnOrder.length - 1)"
            :key="colId"
            class="flex flex-row items-center gap-1"
          >
            <Button size="icon" class="size-6" variant="ghost" :disabled="i === 0" @click="moveHeader(i + 1, i)">
              <MdiArrowUp />
            </Button>
            <Button
              size="icon"
              class="size-6"
              variant="ghost"
              :disabled="i === columnOrder.length - 3"
              @click="moveHeader(i + 1, i + 2)"
            >
              <MdiArrowDown />
            </Button>
            <Checkbox
              :id="colId"
              :model-value="table.getColumn(colId)?.getIsVisible()"
              @update:model-value="toggleHeader(colId)"
            />
            <label class="text-sm" :for="colId"> {{ $t(`items.${camelToSnakeCase(colId)}`) }} </label>
          </div>
        </div>

        <div class="mt-4 flex flex-col gap-2">
          <Label> {{ $t("components.item.view.table.rows_per_page") }} </Label>
          <Select :model-value="pagination.pageSize" @update:model-value="val => table.setPageSize(Number(val))">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="1">1</SelectItem>
              <SelectItem :value="10">12</SelectItem>
              <SelectItem :value="25">24</SelectItem>
              <SelectItem :value="50">48</SelectItem>
              <SelectItem :value="100">96</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </DialogContent>
    </Dialog>
    <BaseCard v-if="props.view === 'table'">
      <div>
        <TableView :table="table" :columns="columns" />
      </div>
      <div v-if="!props.disableControls" class="border-t p-3">
        <DataTableControls :table="table" :pagination="pagination" :data-length="data.length" />
      </div>
    </BaseCard>
    <div v-else>
      <CardView :table="table" />
      <div v-if="!props.disableControls" class="pt-2">
        <DataTableControls :table="table" :pagination="pagination" :data-length="data.length" />
      </div>
    </div>
  </div>
</template>

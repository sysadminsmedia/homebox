<script setup lang="ts" generic="TData, TValue">
  import BaseCard from "@/components/Base/Card.vue";
  import type { ColumnDef, SortingState, VisibilityState, ExpandedState } from "@tanstack/vue-table";
  import {
    FlexRender,
    getCoreRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    getExpandedRowModel,
    useVueTable,
  } from "@tanstack/vue-table";

  import { camelToSnakeCase, valueUpdater } from "@/lib/utils";

  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import Button from "~/components/ui/button/Button.vue";
  import { DialogID, useDialog } from "~/components/ui/dialog-provider/utils";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiTableCog from "~icons/mdi/table-cog";
  import Checkbox from "~/components/ui/checkbox/Checkbox.vue";
  import Label from "~/components/ui/label/Label.vue";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";

  const { openDialog } = useDialog();

  const props = defineProps<{
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
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
    pageSize: defaultPageSize || 10,
  });

  watch(
    () => pagination.value.pageSize,
    newSize => {
      preferences.value.itemsPerTablePage = newSize;
    }
  );

  const table = useVueTable({
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

        <div>{{ $t("components.item.view.table.headers") }}</div>
        <div class="flex flex-col">
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
              <SelectItem :value="10">10</SelectItem>
              <SelectItem :value="25">25</SelectItem>
              <SelectItem :value="50">50</SelectItem>
              <SelectItem :value="100">100</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </DialogContent>
    </Dialog>
    <BaseCard>
      <div>
        <Table class="w-full">
          <TableHeader>
            <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
              <TableHead
                v-for="header in headerGroup.headers"
                :key="header.id"
                class="text-no-transform cursor-pointer bg-secondary text-sm text-secondary-foreground hover:bg-secondary/90"
              >
                <FlexRender
                  v-if="!header.isPlaceholder"
                  :render="header.column.columnDef.header"
                  :props="header.getContext()"
                />
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <template v-if="table.getRowModel().rows?.length">
              <template v-for="row in table.getRowModel().rows" :key="row.id">
                <TableRow :data-state="row.getIsSelected() ? 'selected' : undefined">
                  <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id">
                    <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
                  </TableCell>
                </TableRow>
                <TableRow v-if="row.getIsExpanded()">
                  <TableCell :colspan="row.getAllCells().length">
                    {{ JSON.stringify(row.original) }}
                  </TableCell>
                </TableRow>
              </template>
            </template>
            <template v-else>
              <TableRow>
                <TableCell :colspan="columns.length" class="h-24 text-center"> No results. </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </div>
      <div class="flex items-center justify-end space-x-2 py-4">
        <Button class="size-10 p-0" variant="outline" @click="openDialog(DialogID.ItemTableSettings)">
          <MdiTableCog />
        </Button>
        <div class="flex-1 text-sm text-muted-foreground">
          {{ table.getFilteredSelectedRowModel().rows.length }} of {{ table.getFilteredRowModel().rows.length }} row(s)
          selected.
        </div>
        <div class="flex items-center justify-end space-x-2 py-4">
          <Button variant="outline" size="sm" :disabled="!table.getCanPreviousPage()" @click="table.previousPage()">
            Previous
          </Button>
          <Button variant="outline" size="sm" :disabled="!table.getCanNextPage()" @click="table.nextPage()">
            Next
          </Button>
        </div>
      </div>
    </BaseCard>
  </div>
</template>

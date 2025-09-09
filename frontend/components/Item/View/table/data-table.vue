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

  import { valueUpdater } from "@/lib/utils";

  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import Button from "~/components/ui/button/Button.vue";
  import { DialogID, useDialog } from "~/components/ui/dialog-provider/utils";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import Checkbox from "~/components/ui/checkbox/Checkbox.vue";
  import Label from "~/components/ui/label/Label.vue";

  const { openDialog, closeDialog } = useDialog();

  const props = defineProps<{
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
  }>();

  const {
    value: { tableHeaders: tableHeadersData },
  } = useViewPreferences();
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
    },
  });
</script>

<template>
  <div>
    <!-- <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button variant="outline" class="ml-auto">
            Columns
            <ChevronDown class="ml-2 size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuCheckboxItem
            v-for="column in table.getAllColumns().filter(column => column.getCanHide())"
            :key="column.id"
            class="capitalize"
            :model-value="column.getIsVisible()"
            @update:model-value="
              value => {
                column.toggleVisibility(!!value);
              }
            "
          >
            {{ column.id }}
          </DropdownMenuCheckboxItem>
        </DropdownMenuContent>
      </DropdownMenu> -->
    <Dialog :dialog-id="DialogID.ItemTableSettings">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("components.item.view.table.table_settings") }}</DialogTitle>
        </DialogHeader>

        <div>{{ $t("components.item.view.table.headers") }}</div>
        <div class="flex flex-col">
          <div v-for="(h, i) in headers" :key="h.value" class="flex flex-row items-center gap-1">
            <Button size="icon" class="size-6" variant="ghost" :disabled="i === 0" @click="moveHeader(i, i - 1)">
              <MdiArrowUp />
            </Button>
            <Button
              size="icon"
              class="size-6"
              variant="ghost"
              :disabled="i === headers.length - 1"
              @click="moveHeader(i, i + 1)"
            >
              <MdiArrowDown />
            </Button>
            <Checkbox :id="h.value" :model-value="h.enabled" @update:model-value="toggleHeader(h.value)" />
            <label class="text-sm" :for="h.value"> {{ $t(h.text) }} </label>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <Label> {{ $t("components.item.view.table.rows_per_page") }} </Label>
          <Select :model-value="pagination.rowsPerPage" @update:model-value="pagination.rowsPerPage = Number($event)">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="10">10</SelectItem>
              <SelectItem :value="25">25</SelectItem>
              <SelectItem :value="50">50</SelectItem>
              <SelectItem :value="100">100</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <DialogFooter>
          <Button @click="closeDialog(DialogID.ItemTableSettings)"> {{ $t("global.save") }} </Button>
        </DialogFooter>
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

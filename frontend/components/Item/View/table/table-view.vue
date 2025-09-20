<script setup lang="ts" generic="TValue">
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import DataTableExpandedRow from "./data-table-expanded-row.vue";
  import { FlexRender, type ColumnDef, type Table as TableType } from "@tanstack/vue-table";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";

  defineProps<{
    table: TableType<ItemSummary>;
    columns: ColumnDef<ItemSummary, TValue>[];
  }>();
</script>

<template>
  <Table class="w-full">
    <TableHeader>
      <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
        <TableHead
          v-for="header in headerGroup.headers"
          :key="header.id"
          :class="[
            'text-no-transform cursor-pointer bg-secondary text-sm text-secondary-foreground hover:bg-secondary/90',
            header.column.id === 'select' || header.column.id === 'actions' ? 'w-10 px-3 text-center' : '',
          ]"
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
            <TableCell
              v-for="cell in row.getVisibleCells()"
              :key="cell.id"
              :href="
                cell.column.id !== 'select' && cell.column.id !== 'actions' ? `/item/${row.original.id}` : undefined
              "
              :class="cell.column.id === 'select' || cell.column.id === 'actions' ? 'w-10 px-3' : ''"
              :compact="cell.column.id === 'select' || cell.column.id === 'actions'"
            >
              <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
            </TableCell>
          </TableRow>
          <TableRow v-if="row.getIsExpanded()">
            <TableCell :colspan="row.getAllCells().length">
              <DataTableExpandedRow :item="row.original" />
            </TableCell>
          </TableRow>
        </template>
      </template>
      <template v-else>
        <TableRow>
          <TableCell :colspan="columns.length" class="h-24 text-center">
            <p>{{ $t("items.no_results") }}</p>
          </TableCell>
        </TableRow>
      </template>
    </TableBody>
  </Table>
</template>

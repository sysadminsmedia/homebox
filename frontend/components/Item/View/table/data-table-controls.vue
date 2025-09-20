<script setup lang="ts">
  import type { Table as TableType } from "@tanstack/vue-table";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";

  import MdiTableCog from "~icons/mdi/table-cog";

  import Button from "~/components/ui/button/Button.vue";
  import {
    Pagination,
    PaginationEllipsis,
    PaginationFirst,
    PaginationLast,
    PaginationList,
    PaginationListItem,
  } from "@/components/ui/pagination";
  import { DialogID, useDialog } from "~/components/ui/dialog-provider/utils";
  import type { Pagination as PaginationType } from "../pagination";

  const { openDialog } = useDialog();

  defineProps<{
    table: TableType<ItemSummary>;
    dataLength: number;
    externalPagination?: PaginationType;
  }>();
</script>

<template>
  <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between md:gap-0">
    <div class="order-2 flex items-center gap-2 md:order-1">
      <Button class="size-10 p-0" variant="outline" @click="openDialog(DialogID.ItemTableSettings)">
        <MdiTableCog />
      </Button>
      <div class="text-sm text-muted-foreground">
        {{
          $t("components.item.view.table.selected_rows", {
            selected: table.getFilteredSelectedRowModel().rows.length,
            total: table.getFilteredRowModel().rows.length,
          })
        }}
      </div>
    </div>
    <div class="order-1 flex w-full justify-center md:order-2 md:w-auto">
      <Pagination
        v-slot="{ page }"
        :items-per-page="externalPagination ? externalPagination.pageSize : table.getState().pagination.pageSize"
        :total="externalPagination ? externalPagination.totalSize : dataLength"
        :sibling-count="2"
        :page="externalPagination ? externalPagination.page : table.getState().pagination.pageIndex + 1"
        @update:page="val => (externalPagination ? externalPagination.setPage(val) : table.setPageIndex(val - 1))"
      >
        <PaginationList v-slot="{ items: pageItems }" class="flex items-center gap-1">
          <PaginationFirst
            @click="() => (externalPagination ? externalPagination.setPage(1) : table.setPageIndex(0))"
          />
          <template v-for="(item, index) in pageItems">
            <PaginationListItem v-if="item.type === 'page'" :key="index" :value="item.value" as-child>
              <Button
                class="size-10 p-0"
                :variant="item.value === page ? 'default' : 'outline'"
                @click="
                  () =>
                    externalPagination ? externalPagination.setPage(item.value) : table.setPageIndex(item.value - 1)
                "
              >
                {{ item.value }}
              </Button>
            </PaginationListItem>
            <PaginationEllipsis v-else :key="item.type" :index="index" />
          </template>
          <PaginationLast
            @click="
              () =>
                externalPagination
                  ? externalPagination.setPage(Math.ceil(externalPagination.totalSize / externalPagination.pageSize))
                  : table.setPageIndex(table.getPageCount() - 1)
            "
          />
        </PaginationList>
      </Pagination>
    </div>
  </div>
</template>

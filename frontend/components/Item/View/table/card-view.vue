<script setup lang="ts">
  import ItemCard from "@/components/Item/Card.vue";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";
  import type { Table as TableType } from "@tanstack/vue-table";
  import MdiSelectSearch from "~icons/mdi/select-search";

  defineProps<{
    table: TableType<ItemSummary>;
    locationFlatTree?: FlatTreeItem[];
  }>();
</script>

<template>
  <div v-if="table.getRowModel().rows?.length === 0" class="flex flex-col items-center gap-2">
    <MdiSelectSearch class="size-10" />
    <p>{{ $t("items.no_results") }}</p>
  </div>
  <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
    <ItemCard
      v-for="item in table.getRowModel().rows"
      :key="item.id"
      :item="item.original"
      :location-flat-tree="locationFlatTree"
    />
  </div>
</template>

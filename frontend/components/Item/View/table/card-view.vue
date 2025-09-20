<script setup lang="ts">
  import ItemCard from "@/components/Item/Card.vue";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";
  import type { Table as TableType } from "@tanstack/vue-table";

  defineProps<{
    table: TableType<ItemSummary>;
  }>();
</script>

<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
    <ItemCard v-for="item in table.getRowModel().rows" :key="item.id" :item="item.original" />
    <div class="hidden first:block">{{ $t("components.item.view.selectable.no_items") }}</div>
  </div>
</template>

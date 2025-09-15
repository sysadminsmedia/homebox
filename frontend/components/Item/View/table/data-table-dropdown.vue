<script setup lang="ts">
  import { MoreHorizontal } from "lucide-vue-next";
  import { Button } from "@/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";
  import type { Column, Row } from "@tanstack/vue-table";

  defineProps<{
    item?: ItemSummary;
    multi?: {
      items: Row<ItemSummary>[];
      columns: Column<ItemSummary>[];
    };
  }>();

  defineEmits<{
    (e: "expand"): void;
  }>();

  const download = (items: Row<ItemSummary>[], columns: Column<ItemSummary>[]) => {
    // get enabled columns
    const enabledColumns = columns.filter(c => c.id !== undefined && c.getIsVisible() && c.getCanHide()).map(c => c.id);

    // create CSV header
    const header = enabledColumns.join(",");

    // map each item to a row matching enabled columns order
    const rows = items.map(item =>
      enabledColumns.map(col => String(item.original[col as keyof ItemSummary] ?? "")).join(",")
    );

    const csv = [header, ...rows].join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "items.csv";
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" class="size-8 p-0 hover:bg-primary hover:text-primary-foreground">
        <span class="sr-only">Open menu</span>
        <MoreHorizontal class="size-4" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuLabel>Actions</DropdownMenuLabel>
      <DropdownMenuItem v-if="item" as-child>
        <NuxtLink :to="`/item/${item.id}`" class="hover:underline"> View item </NuxtLink>
      </DropdownMenuItem>
      <DropdownMenuItem v-if="multi" @click="console.log('needs to be implemented')"> View items </DropdownMenuItem>
      <DropdownMenuItem @click="$emit('expand')"> Toggle Expand </DropdownMenuItem>
      <DropdownMenuItem v-if="multi" @click="download(multi.items, multi.columns)">
        Download Table as CSV
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

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
  import { useI18n } from "vue-i18n";

  const { t } = useI18n();

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

  const downloadCsv = (items: Row<ItemSummary>[], columns: Column<ItemSummary>[]) => {
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

  const downloadJson = (items: Row<ItemSummary>[], columns: Column<ItemSummary>[]) => {
    // get enabled columns
    const enabledColumns = columns.filter(c => c.id !== undefined && c.getIsVisible() && c.getCanHide()).map(c => c.id);

    // map each item to an object with only enabled columns
    const data = items.map(item => {
      const obj: Record<string, unknown> = {};
      enabledColumns.forEach(col => {
        obj[col] = item.original[col as keyof ItemSummary] ?? null;
      });
      return obj;
    });

    const exportObj = {
      headers: enabledColumns,
      data,
    };

    const json = JSON.stringify(exportObj, null, 2);
    const blob = new Blob([json], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "items.json";
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" class="size-8 p-0 hover:bg-primary hover:text-primary-foreground">
        <span class="sr-only">{{ t("components.item.view.table.dropdown.open_menu") }}</span>
        <MoreHorizontal class="size-4" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuLabel>{{ t("components.item.view.table.dropdown.actions") }}</DropdownMenuLabel>
      <DropdownMenuItem v-if="item" as-child>
        <NuxtLink :to="`/item/${item.id}`" class="hover:underline">
          {{ t("components.item.view.table.dropdown.view_item") }}
        </NuxtLink>
      </DropdownMenuItem>
      <DropdownMenuItem v-if="multi" @click="console.log('needs to be implemented')">
        {{ t("components.item.view.table.dropdown.view_items") }}
      </DropdownMenuItem>
      <DropdownMenuItem @click="$emit('expand')">
        {{ t("components.item.view.table.dropdown.toggle_expand") }}
      </DropdownMenuItem>
      <DropdownMenuItem v-if="multi" @click="downloadCsv(multi.items, multi.columns)">
        {{ t("components.item.view.table.dropdown.download_csv") }}
      </DropdownMenuItem>
      <DropdownMenuItem v-if="multi" @click="downloadJson(multi.items, multi.columns)">
        {{ t("components.item.view.table.dropdown.download_json") }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

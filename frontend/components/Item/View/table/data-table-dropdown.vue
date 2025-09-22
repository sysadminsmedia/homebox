<script setup lang="ts">
  import { MoreHorizontal } from "lucide-vue-next";
  import { Button } from "@/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuTrigger,
    DropdownMenuSeparator,
  } from "@/components/ui/dropdown-menu";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";
  import type { Column, Row, Table } from "@tanstack/vue-table";
  import { useI18n } from "vue-i18n";
  import { toast } from "~/components/ui/sonner";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";

  const { t } = useI18n();
  const api = useUserApi();
  const confirm = useConfirm();
  const preferences = useViewPreferences();
  const { openDialog } = useDialog();

  const props = defineProps<{
    item?: ItemSummary;
    multi?: {
      items: Row<ItemSummary>[];
      columns: Column<ItemSummary>[];
    };
    view: "table" | "card";
    table: Table<ItemSummary>;
  }>();

  const emit = defineEmits<{
    (e: "expand"): void;
    (e: "refresh"): void;
  }>();

  const resetSelection = () => {
    props.table.resetRowSelection();
    props.table.resetExpanded();
    emit("refresh");
  };

  const openMultiTab = async (items: string[]) => {
    if (!preferences.value.shownMultiTabWarning) {
      // TODO: add warning with link to docs and just improve this
      const { isCanceled } = await confirm.open({
        message: t("components.item.view.table.dropdown.open_multi_tab_warning"),
        href: "https://homebox.software/en/user-guide/tips-tricks#open-multiple-items-in-new-tabs",
      });
      if (isCanceled) {
        return;
      }
      preferences.value.shownMultiTabWarning = true;
    }

    items.forEach(item => window.open(`/item/${item}`, "_blank"));
  };

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

  const deleteItems = async (ids: string[]) => {
    const { isCanceled } = await confirm.open(t("components.item.view.table.dropdown.delete_confirmation"));

    if (isCanceled) {
      return;
    }

    await Promise.allSettled(
      ids.map(id =>
        api.items.delete(id).catch(err => {
          toast.error(t("components.item.view.table.dropdown.error_deleting"));
          console.error(err);
        })
      )
    );

    resetSelection();
  };

  const duplicateItems = async (ids: string[]) => {
    await Promise.allSettled(
      ids.map(id =>
        api.items
          .duplicate(id, {
            copyMaintenance: preferences.value.duplicateSettings.copyMaintenance,
            copyAttachments: preferences.value.duplicateSettings.copyAttachments,
            copyCustomFields: preferences.value.duplicateSettings.copyCustomFields,
            copyPrefix: preferences.value.duplicateSettings.copyPrefixOverride ?? t("items.duplicate.prefix"),
          })
          .catch(err => {
            toast.error(t("components.item.view.table.dropdown.error_duplicating"));
            console.error(err);
          })
      )
    );

    resetSelection();
  };
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button
        :variant="view === 'table' ? 'ghost' : 'outline'"
        class="size-8 p-0 hover:bg-primary hover:text-primary-foreground"
      >
        <span class="sr-only">{{ t("components.item.view.table.dropdown.open_menu") }}</span>
        <MoreHorizontal class="size-4" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuLabel>{{ t("components.item.view.table.dropdown.actions") }}</DropdownMenuLabel>
      <DropdownMenuSeparator />
      <DropdownMenuItem v-if="item" as-child>
        <NuxtLink :to="`/item/${item.id}`" class="hover:underline">
          {{ t("components.item.view.table.dropdown.view_item") }}
        </NuxtLink>
      </DropdownMenuItem>
      <DropdownMenuItem v-if="multi" @click="openMultiTab(multi.items.map(row => row.original.id))">
        {{ t("components.item.view.table.dropdown.view_items") }}
      </DropdownMenuItem>
      <DropdownMenuItem v-if="view === 'table'" @click="$emit('expand')">
        {{ t("components.item.view.table.dropdown.toggle_expand") }}
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <!-- change location -->
      <DropdownMenuItem
        @click="
          openDialog(DialogID.ItemChangeDetails, {
            params: { items: multi ? multi.items.map(row => row.original) : [item!], changeLocation: true },
            onClose: result => {
              if (result) {
                toast.success(t('components.item.view.table.dropdown.change_location_success'));
                resetSelection();
              }
            },
          })
        "
      >
        {{ t("components.item.view.table.dropdown.change_location") }}
      </DropdownMenuItem>
      <!-- change labels -->
      <DropdownMenuItem
        @click="
          openDialog(DialogID.ItemChangeDetails, {
            params: {
              items: multi ? multi.items.map(row => row.original) : [item!],
              addLabels: true,
              removeLabels: true,
            },
            onClose: result => {
              if (result) {
                toast.success(t('components.item.view.table.dropdown.change_labels_success'));
                resetSelection();
              }
            },
          })
        "
      >
        {{ t("components.item.view.table.dropdown.change_labels") }}
      </DropdownMenuItem>
      <!-- maintenance -->
      <DropdownMenuItem
        @click="
          openDialog(DialogID.EditMaintenance, {
            params: { type: 'create', itemId: multi ? multi.items.map(row => row.original.id) : item!.id },
            onClose: result => {
              if (result) {
                toast.success(t('components.item.view.table.dropdown.create_maintenance_success'));
              }
            },
          })
        "
      >
        {{
          multi
            ? t("components.item.view.table.dropdown.create_maintenance_selected")
            : t("components.item.view.table.dropdown.create_maintenance_item")
        }}
      </DropdownMenuItem>
      <!-- duplicate -->
      <DropdownMenuItem @click="duplicateItems(multi ? multi.items.map(row => row.original.id) : [item!.id])">
        {{
          multi
            ? t("components.item.view.table.dropdown.duplicate_selected")
            : t("components.item.view.table.dropdown.duplicate_item")
        }}
      </DropdownMenuItem>
      <!-- delete -->
      <DropdownMenuItem @click="deleteItems(multi ? multi.items.map(row => row.original.id) : [item!.id])">
        {{
          multi
            ? t("components.item.view.table.dropdown.delete_selected")
            : t("components.item.view.table.dropdown.delete_item")
        }}
      </DropdownMenuItem>
      <!-- download -->
      <DropdownMenuSeparator v-if="multi && view === 'table'" />
      <DropdownMenuItem v-if="multi && view === 'table'" @click="downloadCsv(multi.items, multi.columns)">
        {{ t("components.item.view.table.dropdown.download_csv") }}
      </DropdownMenuItem>
      <DropdownMenuItem v-if="multi && view === 'table'" @click="downloadJson(multi.items, multi.columns)">
        {{ t("components.item.view.table.dropdown.download_json") }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

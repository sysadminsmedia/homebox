<template>
  <!-- <Dialog :dialog-id="DialogID.ItemTableSettings">
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
  </Dialog> -->
  <BaseCard>
    <Table class="w-full">
      <TableHeader>
        <TableRow>
          <TableHead
            v-for="h in headers.filter(h => h.enabled)"
            :key="h.value"
            class="text-no-transform cursor-pointer bg-secondary text-sm text-secondary-foreground hover:bg-secondary/90"
            @click="sortBy(h.value)"
          >
            <div
              class="flex items-center gap-1"
              :class="{
                'justify-center': h.align === 'center',
                'justify-start': h.align === 'right',
                'justify-end': h.align === 'left',
              }"
            >
              <template v-if="typeof h === 'string'">{{ h }}</template>
              <template v-else>{{ $t(h.text) }}</template>
              <div
                :data-swap="pagination.descending"
                :class="{ 'opacity-0': sortByProperty !== h.value }"
                class="transition-transform duration-300 data-[swap=true]:rotate-180"
              >
                <MdiArrowUp class="size-5" />
              </div>
            </div>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="(d, i) in data" :key="d.id" class="relative cursor-pointer">
          <TableCell
            v-for="h in headers.filter(h => h.enabled)"
            :key="`${h.value}-${i}`"
            :class="{
              'text-center': h.align === 'center',
              'text-right': h.align === 'right',
              'text-left': h.align === 'left',
            }"
          >
            <template v-if="h.type === 'name'">
              {{ d.name }}
            </template>
            <template v-else-if="h.type === 'price'">
              <Currency :amount="d.purchasePrice" />
            </template>
            <template v-else-if="h.type === 'boolean'">
              <MdiCheck v-if="d.insured" class="inline size-5 text-green-500" />
              <MdiClose v-else class="inline size-5 text-destructive" />
            </template>
            <template v-else-if="h.type === 'location'">
              <NuxtLink v-if="d.location" class="hover:underline" :to="`/location/${d.location.id}`">
                {{ d.location.name }}
              </NuxtLink>
            </template>
            <template v-else-if="h.type === 'date'">
              <DateTime :date="d[h.value]" datetime-type="date" />
            </template>
            <slot v-else :name="cell(h)" v-bind="{ item: d }">
              {{ extractValue(d, h.value) }}
            </slot>
          </TableCell>
          <TableCell class="absolute inset-0">
            <NuxtLink :to="`/item/${d.id}`" class="absolute inset-0">
              <span class="sr-only">{{ $t("components.item.view.table.view_item") }}</span>
            </NuxtLink>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    <div
      class="flex items-center justify-between gap-2 border-t p-3"
      :class="{
        hidden: disableControls,
      }"
    >
      <Button class="size-10 p-0" variant="outline" @click="openDialog(DialogID.ItemTableSettings)">
        <MdiTableCog />
      </Button>
      <Pagination
        v-slot="{ page }"
        :items-per-page="pagination.rowsPerPage"
        :total="props.items.length"
        :sibling-count="2"
        @update:page="pagination.page = $event"
      >
        <PaginationList v-slot="{ items: pageItems }" class="flex items-center gap-1">
          <PaginationFirst />
          <template v-for="(item, index) in pageItems">
            <PaginationListItem v-if="item.type === 'page'" :key="index" :value="item.value" as-child>
              <Button class="size-10 p-0" :variant="item.value === page ? 'default' : 'outline'">
                {{ item.value }}
              </Button>
            </PaginationListItem>
            <PaginationEllipsis v-else :key="item.type" :index="index" />
          </template>
          <PaginationLast />
        </PaginationList>
      </Pagination>
      <Button class="invisible hidden size-10 p-0 md:block">
        <!-- properly centre the pagination buttons -->
      </Button>
    </div>
  </BaseCard>
</template>

<script setup lang="ts">
  import type { TableData, TableHeaderType } from "./Table.types";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiCheck from "~icons/mdi/check";
  import MdiClose from "~icons/mdi/close";
  import MdiTableCog from "~icons/mdi/table-cog";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import {
    Pagination,
    PaginationEllipsis,
    PaginationFirst,
    PaginationLast,
    PaginationList,
    PaginationListItem,
  } from "@/components/ui/pagination";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import BaseCard from "@/components/Base/Card.vue";
  import Currency from "~/components/global/Currency.vue";
  import DateTime from "~/components/global/DateTime.vue";

  const { openDialog, closeDialog } = useDialog();

  type Props = {
    items: ItemSummary[];
    disableControls?: boolean;
  };
  const props = defineProps<Props>();

  const sortByProperty = ref<keyof ItemSummary | "">("");

  const preferences = useViewPreferences();

  const defaultHeaders = [
    { text: "items.asset_id", value: "assetId", enabled: false },
    {
      text: "items.name",
      value: "name",
      enabled: true,
      type: "name",
    },
    { text: "items.quantity", value: "quantity", align: "center", enabled: true },
    { text: "items.insured", value: "insured", align: "center", enabled: true, type: "boolean" },
    { text: "items.purchase_price", value: "purchasePrice", align: "center", enabled: true, type: "price" },
    { text: "items.location", value: "location", align: "center", enabled: false, type: "location" },
    { text: "items.archived", value: "archived", align: "center", enabled: false, type: "boolean" },
    { text: "items.created_at", value: "createdAt", align: "center", enabled: false, type: "date" },
    { text: "items.updated_at", value: "updatedAt", align: "center", enabled: false, type: "date" },
  ] satisfies TableHeaderType[];

  const headers = ref<TableHeaderType[]>(
    (preferences.value.tableHeaders ?? [])
      .concat(defaultHeaders.filter(h => !preferences.value.tableHeaders?.find(h2 => h2.value === h.value)))
      // this is a hack to make sure that any changes to the defaultHeaders are reflected in the preferences
      .map(h => ({
        ...(defaultHeaders.find(h2 => h2.value === h.value) as TableHeaderType),
        enabled: h.enabled,
      }))
  );

  const toggleHeader = (value: string) => {
    const header = headers.value.find(h => h.value === value);
    if (header) {
      header.enabled = !header.enabled; // Toggle the 'enabled' state
    }

    preferences.value.tableHeaders = headers.value;
  };
  const moveHeader = (from: number, to: number) => {
    const header = headers.value[from];
    if (!header) {
      return;
    }
    headers.value.splice(from, 1);
    headers.value.splice(to, 0, header);

    preferences.value.tableHeaders = headers.value;
  };

  const pagination = reactive({
    descending: false,
    page: 1,
    rowsPerPage: preferences.value.itemsPerTablePage,
    rowsNumber: 0,
  });

  watch(
    () => pagination.rowsPerPage,
    newRowsPerPage => {
      preferences.value.itemsPerTablePage = newRowsPerPage;
    }
  );

  function sortBy(property: keyof ItemSummary) {
    if (sortByProperty.value === property) {
      pagination.descending = !pagination.descending;
    } else {
      pagination.descending = false;
    }
    sortByProperty.value = property;
  }

  function extractSortable(item: ItemSummary, property: keyof ItemSummary): string | number | boolean {
    const value = item[property];
    if (typeof value === "string") {
      // Try to parse number
      const parsed = Number(value);
      if (!isNaN(parsed)) {
        return parsed;
      }

      return value.toLowerCase();
    }

    if (typeof value !== "number" && typeof value !== "boolean") {
      return "";
    }

    return value;
  }

  function itemSort(a: ItemSummary, b: ItemSummary) {
    if (!sortByProperty.value) {
      return 0;
    }

    const aVal = extractSortable(a, sortByProperty.value);
    const bVal = extractSortable(b, sortByProperty.value);

    if (typeof aVal === "string" && typeof bVal === "string") {
      return aVal.localeCompare(bVal, undefined, { numeric: true, sensitivity: "base" });
    }

    if (aVal < bVal) {
      return -1;
    }
    if (aVal > bVal) {
      return 1;
    }
    return 0;
  }

  const data = computed<TableData[]>(() => {
    // sort by property
    let data = [...props.items].sort(itemSort);

    // sort descending
    if (pagination.descending) {
      data.reverse();
    }

    // paginate
    const start = (pagination.page - 1) * pagination.rowsPerPage;
    const end = start + pagination.rowsPerPage;
    data = data.slice(start, end);
    return data;
  });

  function extractValue(data: TableData, value: string) {
    const parts = value.split(".");
    let current = data;
    for (const part of parts) {
      current = current[part];
    }
    return current;
  }

  function cell(h: TableHeaderType) {
    return `cell-${h.value.replace(".", "_")}`;
  }
</script>

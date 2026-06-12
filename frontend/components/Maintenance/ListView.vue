<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { DialogRoot } from "reka-ui";
  import type { EntitySummary, MaintenanceEntry, MaintenanceEntryWithDetails } from "~~/lib/api/types/data-contracts";
  import { MaintenanceFilterStatus } from "~~/lib/api/types/data-contracts";
  import type { StatsFormat } from "~~/components/global/StatCard/types";
  import MdiCheck from "~icons/mdi/check";
  import MdiDelete from "~icons/mdi/delete";
  import MdiEdit from "~icons/mdi/edit";
  import MdiCalendar from "~icons/mdi/calendar";
  import MdiRepeat from "~icons/mdi/repeat";
  import MdiPlus from "~icons/mdi/plus";
  import MdiAlertCircle from "~icons/mdi/alert-circle";
  import MdiWrenchClock from "~icons/mdi/wrench-clock";
  import MdiContentDuplicate from "~icons/mdi/content-duplicate";
  import MaintenanceEditModal from "~~/components/Maintenance/EditModal.vue";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import { Badge } from "@/components/ui/badge";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import { DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import StatCard from "~/components/global/StatCard/StatCard.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import Currency from "~/components/global/Currency.vue";
  import Markdown from "~/components/global/Markdown.vue";
  import ItemSelector from "~/components/Item/Selector.vue";
  import { toast } from "@/components/ui/sonner";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { toDateOnlyString } from "~/lib/datelib/dateOnly";
  import { DialogID } from "../ui/dialog-provider/utils";
  import { useDebounceFn } from "@vueuse/core";

  const maintenanceFilterStatus = ref(MaintenanceFilterStatus.MaintenanceFilterStatusScheduled);

  const api = useUserApi();
  const { t } = useI18n();
  const confirm = useConfirm();
  const { openDialog } = useDialog();

  const props = defineProps({
    currentItemId: {
      type: String,
      default: undefined,
    },
  });

  const itemPickerOpen = ref(false);
  const selectedItem = ref<EntitySummary | null>(null);
  const itemSearch = ref("");
  const availableItems = ref<EntitySummary[]>([]);
  const isLoadingItems = ref(false);
  let itemSearchRequestId = 0;

  const { data: maintenanceDataList, refresh: refreshList } = useAsyncData(
    async () => {
      const { data } =
        props.currentItemId !== undefined
          ? await api.items.maintenance.getLog(props.currentItemId, { status: maintenanceFilterStatus.value })
          : await api.maintenance.getAll({ status: maintenanceFilterStatus.value });
      return data;
    },
    {
      watch: [maintenanceFilterStatus],
    }
  );

  const stats = computed(() => {
    if (!maintenanceDataList.value) return [];

    const count = maintenanceDataList.value ? maintenanceDataList.value.length || 0 : 0;
    let total = 0;
    maintenanceDataList.value.forEach(item => {
      total += parseFloat(item.cost);
    });

    const average = count > 0 ? total / count : 0;

    return [
      {
        id: "count",
        title: t("maintenance.total_entries"),
        value: count,
        type: "number" as StatsFormat,
      },
      {
        id: "total",
        title: t("maintenance.total_cost"),
        value: total,
        type: "currency" as StatsFormat,
      },
      {
        id: "average",
        title: t("maintenance.monthly_average"),
        value: average,
        type: "currency" as StatsFormat,
      },
    ];
  });

  async function deleteEntry(id: string) {
    const result = await confirm.open(t("maintenance.modal.delete_confirmation"));
    if (result.isCanceled) {
      return;
    }

    const { error } = await api.maintenance.delete(id);

    if (error) {
      toast.error(t("maintenance.toast.failed_to_delete"));
      return;
    }
    refreshList();
  }

  async function completeEntry(maintenanceEntry: MaintenanceEntry) {
    const { error } = await api.maintenance.update(maintenanceEntry.id, {
      name: maintenanceEntry.name,
      completedDate: toDateOnlyString(new Date()),
      scheduledDate: (maintenanceEntry.scheduledDate as string) ?? "",
      planID: maintenanceEntry.planID,
      description: maintenanceEntry.description,
      cost: maintenanceEntry.cost,
    });
    if (error) {
      toast.error(t("maintenance.toast.failed_to_update"));
    }
    refreshList();
  }

  function hasRecurringPlan(entry: MaintenanceEntry | MaintenanceEntryWithDetails): boolean {
    if (!entry.planID) {
      return false;
    }

    return entry.planID !== "00000000-0000-0000-0000-000000000000";
  }

  async function searchItems(query: string) {
    const requestId = ++itemSearchRequestId;
    isLoadingItems.value = true;
    const { data, error } = await api.items.getAll({
      q: query,
      page: 1,
      pageSize: 100,
    });
    isLoadingItems.value = false;

    if (requestId !== itemSearchRequestId) {
      return false;
    }

    if (error || !data) {
      return false;
    }

    availableItems.value = data.items;
    return true;
  }

  const debouncedSearchItems = useDebounceFn((query: string) => {
    void searchItems(query);
  }, 300);

  async function loadInitialItems() {
    if (availableItems.value.length > 0) {
      return true;
    }

    return searchItems("");
  }

  watch(
    itemSearch,
    query => {
      if (!itemPickerOpen.value) {
        return;
      }

      debouncedSearchItems(query);
    },
    { immediate: false }
  );

  function openMaintenanceModalForCurrentItem() {
    openDialog(DialogID.EditMaintenance, {
      params: { type: "create", itemId: props.currentItemId },
      onClose: result => {
        if (result) {
          refreshList();
        }
      },
    });
  }

  function openItemPickerModal() {
    selectedItem.value = null;
    itemSearch.value = "";
    itemPickerOpen.value = true;
    void loadInitialItems();
  }

  function openMaintenanceModalForSelectedItem() {
    if (!selectedItem.value) {
      return;
    }

    itemPickerOpen.value = false;
    openDialog(DialogID.EditMaintenance, {
      params: { type: "create", itemId: selectedItem.value.id },
      onClose: result => {
        if (result) {
          refreshList();
        }
      },
    });
  }
</script>

<template>
  <section class="space-y-6">
    <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
      <StatCard v-for="stat in stats" :key="stat.id" :title="stat.title" :value="stat.value" :type="stat.type" />
    </div>
    <div class="flex">
      <ButtonGroup>
        <Button
          size="sm"
          :variant="
            maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusScheduled ? 'default' : 'outline'
          "
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusScheduled"
        >
          {{ $t("maintenance.filter.scheduled") }}
        </Button>
        <Button
          size="sm"
          :variant="
            maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusOverdue ? 'default' : 'outline'
          "
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusOverdue"
        >
          {{ $t("maintenance.filter.overdue") }}
        </Button>
        <Button
          size="sm"
          :variant="
            maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusCompleted ? 'default' : 'outline'
          "
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusCompleted"
        >
          {{ $t("maintenance.filter.completed") }}
        </Button>
        <Button
          size="sm"
          :variant="
            maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusBoth ? 'default' : 'outline'
          "
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusBoth"
        >
          {{ $t("maintenance.filter.both") }}
        </Button>
      </ButtonGroup>
      <Button
        class="ml-auto"
        size="sm"
        @click="props.currentItemId ? openMaintenanceModalForCurrentItem() : openItemPickerModal()"
      >
        <MdiPlus />
        {{ $t("maintenance.list.new") }}
      </Button>
    </div>
  </section>
  <section>
    <!-- begin -->
    <MaintenanceEditModal ref="maintenanceEditModal" @changed="refreshList" />
    <DialogRoot :open="itemPickerOpen" @update:open="itemPickerOpen = $event">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("maintenance.list.select_item") }}</DialogTitle>
        </DialogHeader>
        <div class="py-2">
          <ItemSelector
            v-model="selectedItem"
            v-model:search="itemSearch"
            :items="availableItems"
            item-text="name"
            item-value="id"
            :is-loading="isLoadingItems"
            :label="$t('global.items')"
            :trigger-search="loadInitialItems"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="itemPickerOpen = false">
            {{ $t("global.cancel") }}
          </Button>
          <Button :disabled="!selectedItem" @click="openMaintenanceModalForSelectedItem">
            {{ $t("global.confirm") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </DialogRoot>
    <div class="container space-y-6">
      <BaseCard v-for="e in maintenanceDataList" :key="e.id">
        <BaseSectionHeader class="border-b p-6">
          <span class="mb-2">
            <span v-if="!props.currentItemId">
              <NuxtLink class="hover:underline" :to="`/item/${(e as MaintenanceEntryWithDetails).itemID}/maintenance`">
                {{ (e as MaintenanceEntryWithDetails).itemName }}
              </NuxtLink>
              -
            </span>
            {{ e.name }}
          </span>
          <template #description>
            <div class="flex flex-wrap gap-2">
              <Badge v-if="validDate(e.completedDate)" variant="outline">
                <MdiCheck class="mr-2" />
                <DateTime :date="e.completedDate" format="human" datetime-type="date" />
                <MdiRepeat v-if="hasRecurringPlan(e)" class="ml-2" />
              </Badge>
              <Badge v-else-if="e.isOverdue" variant="destructive">
                <MdiAlertCircle class="mr-2" />
                <span v-if="validDate(e.scheduledDate)">
                  {{ $t("maintenance.list.overdue_since") }}
                  <DateTime :date="e.scheduledDate" format="human" datetime-type="date" />
                </span>
                <span v-else>
                  {{ $t("maintenance.list.overdue") }}
                </span>
              </Badge>
              <Badge v-else-if="validDate(e.scheduledDate)" variant="outline">
                <MdiCalendar class="mr-2" />
                <DateTime :date="e.scheduledDate" format="human" datetime-type="date" />
                <MdiRepeat v-if="hasRecurringPlan(e)" class="ml-2" />
              </Badge>
              <TooltipProvider :delay-duration="0">
                <Tooltip>
                  <TooltipTrigger>
                    <Badge>
                      <Currency :amount="e.cost" />
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent> {{ $t("maintenance.modal.cost") }} </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </template>
        </BaseSectionHeader>
        <div :class="{ 'p-6': e.description }">
          <Markdown :source="e.description" />
        </div>
        <ButtonGroup class="flex flex-wrap justify-end p-4">
          <Button
            size="sm"
            @click="
              openDialog(DialogID.EditMaintenance, {
                params: { type: 'update', maintenanceEntry: e, itemId: props.currentItemId ?? undefined },
                onClose: result => {
                  if (result) {
                    refreshList();
                  }
                },
              })
            "
          >
            <MdiEdit />
            {{ $t("maintenance.list.edit") }}
          </Button>
          <Button v-if="!validDate(e.completedDate)" size="sm" variant="outline" @click="completeEntry(e)">
            <MdiCheck />
            {{ $t("maintenance.list.complete") }}
          </Button>
          <Button
            size="sm"
            variant="outline"
            @click="
              openDialog(DialogID.EditMaintenance, {
                params: { type: 'duplicate', maintenanceEntry: e, itemId: props.currentItemId! },
                onClose: result => {
                  if (result) {
                    refreshList();
                  }
                },
              })
            "
          >
            <MdiContentDuplicate />
            {{ $t("maintenance.list.duplicate") }}
          </Button>
          <Button size="sm" variant="destructive" @click="deleteEntry(e.id)">
            <MdiDelete />
            {{ $t("maintenance.list.delete") }}
          </Button>
        </ButtonGroup>
      </BaseCard>
      <div v-if="props.currentItemId" class="hidden first:block">
        <button
          type="button"
          class="relative block w-full rounded-lg border-2 border-dashed p-12 text-center"
          @click="
            openDialog(DialogID.EditMaintenance, {
              params: { type: 'create', itemId: props.currentItemId },
              onClose: result => {
                if (result) {
                  refreshList();
                }
              },
            })
          "
        >
          <MdiWrenchClock class="inline size-16" />
          <span class="mt-2 block text-sm font-medium text-foreground">
            {{ $t("maintenance.list.create_first") }}
          </span>
        </button>
      </div>
    </div>
  </section>
</template>

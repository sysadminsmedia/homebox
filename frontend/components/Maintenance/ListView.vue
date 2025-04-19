<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import type { MaintenanceEntryWithDetails } from "~~/lib/api/types/data-contracts";
  import { MaintenanceFilterStatus } from "~~/lib/api/types/data-contracts";
  import type { StatsFormat } from "~~/components/global/StatCard/types";
  import MdiCheck from "~icons/mdi/check";
  import MdiDelete from "~icons/mdi/delete";
  import MdiEdit from "~icons/mdi/edit";
  import MdiCalendar from "~icons/mdi/calendar";
  import MdiPlus from "~icons/mdi/plus";
  import MdiWrenchClock from "~icons/mdi/wrench-clock";
  import MdiContentDuplicate from "~icons/mdi/content-duplicate";
  import MaintenanceEditModal from "~~/components/Maintenance/EditModal.vue";
  import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";
  import { Badge } from "@/components/ui/badge";
  import { ButtonGroup, Button } from "@/components/ui/button";

  const maintenanceFilterStatus = ref(MaintenanceFilterStatus.MaintenanceFilterStatusScheduled);
  const maintenanceEditModal = ref<InstanceType<typeof MaintenanceEditModal>>();

  const api = useUserApi();
  const { t } = useI18n();

  const props = defineProps({
    currentItemId: {
      type: String,
      default: undefined,
    },
  });

  const { data: maintenanceDataList, refresh: refreshList } = useAsyncData(
    async () => {
      const { data } =
        props.currentItemId !== undefined
          ? await api.items.maintenance.getLog(props.currentItemId, { status: maintenanceFilterStatus.value })
          : await api.maintenance.getAll({ status: maintenanceFilterStatus.value });
      console.log(data);
      return data;
    },
    {
      watch: [maintenanceFilterStatus],
    }
  );

  const stats = computed(() => {
    console.log(maintenanceDataList);
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
        v-if="props.currentItemId"
        class="ml-auto"
        size="sm"
        @click="maintenanceEditModal?.openCreateModal(props.currentItemId)"
      >
        <MdiPlus />
        {{ $t("maintenance.list.new") }}
      </Button>
    </div>
  </section>
  <section>
    <!-- begin -->
    <MaintenanceEditModal ref="maintenanceEditModal" @changed="refreshList"></MaintenanceEditModal>
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
              </Badge>
              <Badge v-else-if="validDate(e.scheduledDate)" variant="outline">
                <MdiCalendar class="mr-2" />
                <DateTime :date="e.scheduledDate" format="human" datetime-type="date" />
              </Badge>
              <TooltipProvider :delay-duration="0">
                <Tooltip>
                  <TooltipTrigger>
                    <Badge>
                      <Currency :amount="e.cost" />
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent> Cost </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </template>
        </BaseSectionHeader>
        <div :class="{ 'p-6': e.description }">
          <Markdown :source="e.description" />
        </div>
        <ButtonGroup class="flex flex-wrap justify-end p-4">
          <Button size="sm" @click="maintenanceEditModal?.openUpdateModal(e)">
            <MdiEdit />
            {{ $t("maintenance.list.edit") }}
          </Button>
          <Button
            v-if="!validDate(e.completedDate)"
            size="sm"
            variant="outline"
            @click="maintenanceEditModal?.complete(e)"
          >
            <MdiCheck />
            {{ $t("maintenance.list.complete") }}
          </Button>
          <Button size="sm" variant="outline" @click="maintenanceEditModal?.duplicate(e, e.itemID)">
            <MdiContentDuplicate />
            {{ $t("maintenance.list.duplicate") }}
          </Button>
          <Button size="sm" variant="destructive" @click="maintenanceEditModal?.deleteEntry(e.id)">
            <MdiDelete />
            {{ $t("maintenance.list.delete") }}
          </Button>
        </ButtonGroup>
      </BaseCard>
      <div v-if="props.currentItemId" class="hidden first:block">
        <button
          type="button"
          class="relative block w-full rounded-lg border-2 border-dashed p-12 text-center"
          @click="maintenanceEditModal?.openCreateModal(props.currentItemId)"
        >
          <MdiWrenchClock class="inline size-16" />
          <span class="mt-2 block text-sm font-medium text-gray-900"> {{ $t("maintenance.list.create_first") }} </span>
        </button>
      </div>
    </div>
  </section>
</template>

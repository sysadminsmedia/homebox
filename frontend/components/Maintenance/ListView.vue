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

  const { data: maintenanceDataList, refresh: refreshList } = useAsyncData<MaintenanceEntryWithDetails[]>(
    async () => {
      const { data } =
        props.currentItemId !== undefined
          ? await api.items.maintenance.getLog(props.currentItemId, { status: maintenanceFilterStatus.value })
          : await api.maintenance.getAll({ status: maintenanceFilterStatus.value });
      console.log(data);
      return data as MaintenanceEntryWithDetails[];
    },
    {
      watch: maintenanceFilterStatus,
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
      <StatCard
        v-for="stat in stats"
        :key="stat.id"
        class="stats border-l-primary block shadow-xl"
        :title="stat.title"
        :value="stat.value"
        :type="stat.type"
      />
    </div>
    <div class="flex">
      <div class="btn-group">
        <BaseButton
          size="sm"
          :class="`${maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusScheduled ? 'btn-active' : ''}`"
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusScheduled"
        >
          {{ $t("maintenance.filter.scheduled") }}
        </BaseButton>
        <BaseButton
          size="sm"
          :class="`${maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusCompleted ? 'btn-active' : ''}`"
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusCompleted"
        >
          {{ $t("maintenance.filter.completed") }}
        </BaseButton>
        <BaseButton
          size="sm"
          :class="`${maintenanceFilterStatus == MaintenanceFilterStatus.MaintenanceFilterStatusBoth ? 'btn-active' : ''}`"
          @click="maintenanceFilterStatus = MaintenanceFilterStatus.MaintenanceFilterStatusBoth"
        >
          {{ $t("maintenance.filter.both") }}
        </BaseButton>
      </div>
      <BaseButton
        v-if="props.currentItemId"
        class="ml-auto"
        size="sm"
        @click="maintenanceEditModal?.openCreateModal(props.currentItemId)"
      >
        <template #icon>
          <MdiPlus />
        </template>
        {{ $t("maintenance.list.new") }}
      </BaseButton>
    </div>
  </section>
  <section>
    <!-- begin -->
    <MaintenanceEditModal ref="maintenanceEditModal" @changed="refreshList"></MaintenanceEditModal>
    <div class="container space-y-6">
      <BaseCard v-for="e in maintenanceDataList" :key="e.id">
        <BaseSectionHeader class="border-b border-b-gray-300 p-6">
          <span class="text-base-content">
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
              <div v-if="validDate(e.completedDate)" class="badge p-3">
                <MdiCheck class="mr-2" />
                <DateTime :date="e.completedDate" format="human" datetime-type="date" />
              </div>
              <div v-else-if="validDate(e.scheduledDate)" class="badge p-3">
                <MdiCalendar class="mr-2" />
                <DateTime :date="e.scheduledDate" format="human" datetime-type="date" />
              </div>
              <div class="tooltip tooltip-primary" data-tip="Cost">
                <div class="badge badge-primary p-3">
                  <Currency :amount="e.cost" />
                </div>
              </div>
            </div>
          </template>
        </BaseSectionHeader>
        <div class="p-6">
          <Markdown :source="e.description" />
        </div>
        <div class="flex flex-wrap justify-end gap-1 p-4">
          <BaseButton size="sm" @click="maintenanceEditModal?.openUpdateModal(e)">
            <template #icon>
              <MdiEdit />
            </template>
            {{ $t("maintenance.list.edit") }}
          </BaseButton>
          <BaseButton v-if="!validDate(e.completedDate)" size="sm" @click="maintenanceEditModal?.complete(e)">
            <template #icon>
              <MdiCheck />
            </template>
            {{ $t("maintenance.list.complete") }}
          </BaseButton>
          <BaseButton size="sm" @click="maintenanceEditModal?.duplicate(e, e.itemID)">
            <template #icon>
              <MdiContentDuplicate />
            </template>
            {{ $t("maintenance.list.duplicate") }}
          </BaseButton>
          <BaseButton size="sm" class="btn-error" @click="maintenanceEditModal?.deleteEntry(e.id)">
            <template #icon>
              <MdiDelete />
            </template>
            {{ $t("maintenance.list.delete") }}
          </BaseButton>
        </div>
      </BaseCard>
      <div v-if="props.currentItemId" class="hidden first:block">
        <button
          type="button"
          class="border-base-content relative block w-full rounded-lg border-2 border-dashed p-12 text-center"
          @click="maintenanceEditModal?.openCreateModal(props.currentItemId)"
        >
          <MdiWrenchClock class="inline size-16" />
          <span class="mt-2 block text-sm font-medium text-gray-900"> {{ $t("maintenance.list.create_first") }} </span>
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import type { StatsFormat } from "~~/components/global/StatCard/types";
  import type { ItemOut } from "~~/lib/api/types/data-contracts";
  import MdiPlus from "~icons/mdi/plus";
  import MdiCheck from "~icons/mdi/check";
  import MdiDelete from "~icons/mdi/delete";
  import MdiEdit from "~icons/mdi/edit";
  import MdiCalendar from "~icons/mdi/calendar";
  import MdiWrenchClock from "~icons/mdi/wrench-clock";
  import MaintenanceEditModal from "~~/components/Maintenance/EditModal.vue";

  const { t } = useI18n();
  const props = defineProps<{
    item: ItemOut;
  }>();

  const api = useUserApi();
  const toast = useNotifier();

  const scheduled = ref(true);

  const maintenanceEditModal = ref<InstanceType<typeof MaintenanceEditModal>>();

  watch(
    () => scheduled.value,
    () => {
      refreshLog();
    }
  );

  const { data: log, refresh: refreshLog } = useAsyncData(async () => {
    const { data } = await api.items.maintenance.getLog(props.item.id, {
      scheduled: scheduled.value,
      completed: !scheduled.value,
    });
    return data;
  });

  const count = computed(() => {
    if (!log.value) return 0;
    return log.value.entries.length;
  });
  const stats = computed(() => {
    if (!log.value) return [];

    return [
      {
        id: "count",
        title: t("maintenances.total_entries"),
        value: count.value || 0,
        type: "number" as StatsFormat,
      },
      {
        id: "total",
        title: t("maintenances.total_cost"),
        value: log.value.costTotal || 0,
        type: "currency" as StatsFormat,
      },
      {
        id: "average",
        title: t("maintenances.monthly_average"),
        value: log.value.costAverage || 0,
        type: "currency" as StatsFormat,
      },
    ];
  });
</script>

<template>
  <div v-if="log">
    <MaintenanceEditModal ref="maintenanceEditModal" @changed="refreshLog"></MaintenanceEditModal>

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
          <button class="btn btn-sm" :class="`${scheduled ? 'btn-active' : ''}`" @click="scheduled = true">
            {{ $t("maintenances.filter.scheduled") }}
          </button>
          <button class="btn btn-sm" :class="`${scheduled ? '' : 'btn-active'}`" @click="scheduled = false">
            {{ $t("maintenances.filter.completed") }}
          </button>
        </div>
        <BaseButton class="ml-auto" size="sm" @click="maintenanceEditModal?.openCreateModal(props.item.id)">
          <template #icon>
            <MdiPlus />
          </template>
          {{ $t("maintenances.list.new") }}
        </BaseButton>
      </div>
      <div class="container space-y-6">
        <BaseCard v-for="e in log.entries" :key="e.id">
          <BaseSectionHeader class="border-b border-b-gray-300 p-6">
            <span class="text-base-content">
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
          <div class="flex justify-end gap-1 p-4">
            <BaseButton size="sm" @click="maintenanceEditModal?.openUpdateModal(e)">
              <template #icon>
                <MdiEdit />
              </template>
              {{ $t("maintenances.list.edit") }}
            </BaseButton>
            <BaseButton size="sm" @click="maintenanceEditModal?.deleteEntry(e.id)">
              <template #icon>
                <MdiDelete />
              </template>
              {{ $t("maintenances.list.delete") }}
            </BaseButton>
          </div>
        </BaseCard>
        <div class="hidden first:block">
          <button
            type="button"
            class="border-base-content relative block w-full rounded-lg border-2 border-dashed p-12 text-center"
            @click="maintenanceEditModal?.openCreateModal(props.item.id)"
          >
            <MdiWrenchClock class="inline size-16" />
            <span class="mt-2 block text-sm font-medium text-gray-900"> {{ $t("maintenances.list.create_first") }} </span>
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

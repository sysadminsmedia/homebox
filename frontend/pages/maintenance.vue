<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import type { StatsFormat } from "~~/components/global/StatCard/types";
  import { MaintenanceFilterStatus } from "~~/lib/api/types/data-contracts";
  import MdiCheck from "~icons/mdi/check";
  import MdiDelete from "~icons/mdi/delete";
  import MdiEdit from "~icons/mdi/edit";
  import MdiCalendar from "~icons/mdi/calendar";
  import MaintenanceEditModal from "~~/components/Maintenance/EditModal.vue";

  const { t } = useI18n();

  const api = useUserApi();

  const maintenanceFilter = ref(MaintenanceFilterStatus.MaintenanceFilterStatusScheduled);
  const maintenanceEditModal = ref<InstanceType<typeof MaintenanceEditModal>>();

  const { data: maintenanceData, refresh: refreshList } = useAsyncData(
    async () => {
      const { data } = await api.maintenance.getAll({ status: maintenanceFilter.value });
      console.log(data);
      return data;
    },
    {
      watch: [maintenanceFilter],
    }
  );

  const stats = computed(() => {
    if (!maintenanceData.value) return [];

    return [
      {
        id: "count",
        title: t("maintenance.total_entries"),
        value: maintenanceData.value ? maintenanceData.value.length || 0 : 0,
        type: "number" as StatsFormat,
      },
    ];
  });
</script>

<template>
  <div>
    <BaseContainer class="mb-6 flex flex-col gap-8">
      <BaseSectionHeader> {{ $t("menu.maintenance") }} </BaseSectionHeader>
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
            <button
              class="btn btn-sm"
              :class="`${maintenanceFilter == MaintenanceFilterStatus.MaintenanceFilterStatusScheduled ? 'btn-active' : ''}`"
              @click="maintenanceFilter = MaintenanceFilterStatus.MaintenanceFilterStatusScheduled"
            >
              {{ $t("maintenance.filter.scheduled") }}
            </button>
            <button
              class="btn btn-sm"
              :class="`${maintenanceFilter == MaintenanceFilterStatus.MaintenanceFilterStatusCompleted ? 'btn-active' : ''}`"
              @click="maintenanceFilter = MaintenanceFilterStatus.MaintenanceFilterStatusCompleted"
            >
              {{ $t("maintenance.filter.completed") }}
            </button>
            <button
              class="btn btn-sm"
              :class="`${maintenanceFilter == MaintenanceFilterStatus.MaintenanceFilterStatusBoth ? 'btn-active' : ''}`"
              @click="maintenanceFilter = MaintenanceFilterStatus.MaintenanceFilterStatusBoth"
            >
              {{ $t("maintenance.filter.both") }}
            </button>
          </div>
        </div>
      </section>
      <section>
        <!-- begin -->
        <MaintenanceEditModal ref="maintenanceEditModal" @changed="refreshList"></MaintenanceEditModal>
        <div class="container space-y-6">
          <BaseCard v-for="e in maintenanceData" :key="e.id">
            <BaseSectionHeader class="border-b border-b-gray-300 p-6">
              <span class="text-base-content">
                <NuxtLink class="hover:underline" :to="`/item/${e.itemID}`">
                  {{ e.itemName }}
                </NuxtLink>
                -
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
                {{ $t("maintenance.list.edit") }}
              </BaseButton>
              <BaseButton size="sm" @click="maintenanceEditModal?.deleteEntry(e.id)">
                <template #icon>
                  <MdiDelete />
                </template>
                {{ $t("maintenance.list.delete") }}
              </BaseButton>
            </div>
          </BaseCard>
        </div>
      </section>
    </BaseContainer>
  </div>
</template>

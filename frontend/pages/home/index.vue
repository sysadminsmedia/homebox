<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { statCardData } from "./statistics";
  import { itemsTable } from "./table";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import { ItemSummaryHeaders } from "~/components/Item/View/Table.types";
  
  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "HomeBox | " + t("menu.home"),
  });

  const api = useUserApi();
  const breakpoints = useBreakpoints();

  const locationStore = useLocationStore();
  const locations = computed(() => locationStore.parentLocations);

  const labelsStore = useLabelStore();
  const labels = computed(() => labelsStore.labels);

  const itemTable = itemsTable(api);
  const stats = statCardData(api);
</script>

<template>
  <div>
    <BaseContainer class="flex flex-col gap-4">
      <section>
        <Subtitle> {{ $t("home.quick_statistics") }} </Subtitle>
        <div class="grid grid-cols-2 gap-2 md:grid-cols-4 md:gap-6">
          <StatCard v-for="(stat, i) in stats" :key="i" :title="stat.label" :value="stat.value" :type="stat.type" />
        </div>
      </section>

      <section>
        <Subtitle> {{ $t("home.recently_added") }} </Subtitle>

        <p v-if="itemTable.items.length === 0" class="ml-2 text-sm">{{ $t("items.no_results") }}</p>
        <BaseCard v-else-if="breakpoints.lg">
          <ItemViewTable :items="itemTable.items" :default-table-headers="ItemSummaryHeaders" disable-controls />
        </BaseCard>
        <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <ItemCard v-for="item in itemTable.items" :key="item.id" :item="item" />
        </div>
      </section>

      <section>
        <Subtitle> {{ $t("home.storage_locations") }} </Subtitle>
        <p v-if="locations.length === 0" class="ml-2 text-sm">{{ $t("locations.no_results") }}</p>
        <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          <LocationCard v-for="location in locations" :key="location.id" :location="location" />
        </div>
      </section>

      <section>
        <Subtitle> {{ $t("home.labels") }} </Subtitle>
        <p v-if="labels.length === 0" class="ml-2 text-sm">{{ $t("labels.no_results") }}</p>
        <div v-else class="flex flex-wrap gap-4">
          <LabelChip v-for="label in labels" :key="label.id" size="lg" :label="label" class="shadow-md" />
        </div>
      </section>
    </BaseContainer>
  </div>
</template>

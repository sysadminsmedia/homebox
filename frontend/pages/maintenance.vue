<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import MaintenanceListView from "~/components/Maintenance/ListView.vue";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "HomeBox | " + t("menu.maintenance"),
  });
</script>

<template>
  <div>
    <BaseContainer class="flex flex-col gap-4">
      <BaseSectionHeader> {{ $t("menu.maintenance") }} </BaseSectionHeader>
      <MaintenanceListView />
    </BaseContainer>
  </div>
</template>

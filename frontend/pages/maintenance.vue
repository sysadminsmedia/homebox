<script setup lang="ts">
  import { useI18n } from "vue-i18n";

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

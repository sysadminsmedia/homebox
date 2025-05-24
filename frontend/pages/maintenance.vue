<script setup lang="ts">
  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Maintenance",
  });
</script>

<template>
  <div>
    <BaseContainer class="flex flex-col gap-4">
      <BaseSectionHeader> {{ $t("menu.maintenance") }} </BaseSectionHeader>
      <MaintenanceListView></MaintenanceListView>
    </BaseContainer>
  </div>
</template>

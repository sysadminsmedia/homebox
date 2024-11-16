<script setup lang="ts"></script>

<template>
  <div>
    <BaseContainer class="mb-6 flex flex-col gap-8">
      <BaseSectionHeader> {{ $t("menu.maintenance") }} </BaseSectionHeader>
      <MaintenanceListView></MaintenanceListView>
    </BaseContainer>
  </div>
</template>

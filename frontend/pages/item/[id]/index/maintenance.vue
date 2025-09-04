<script setup lang="ts">
  import type { ItemOut } from "~~/lib/api/types/data-contracts";
  import BaseContainer from "@/components/Base/Container.vue";
  import MaintenanceListView from "~/components/Maintenance/ListView.vue";

  const props = defineProps<{
    item: ItemOut;
  }>();
</script>

<template>
  <BaseContainer class="flex flex-col gap-8">
    <MaintenanceListView :current-item-id="props.item.id" />
  </BaseContainer>
</template>

<script setup lang="ts">
  import type { TreeItem } from "~~/lib/api/types/data-contracts";

  type Props = {
    locs: TreeItem[];
    treeId: string;
  };

  defineProps<Props>();
</script>

<template>
  <div class="root border-2 p-4">
    <p v-if="locs.length === 0" class="text-center text-sm">
      No locations available. Add new locations through the
      <span class="link-primary">Create</span> button on the navigation bar.
    </p>
    <LocationTreeNode v-for="item in locs" :key="item.id" :item="item" :tree-id="treeId" />
  </div>
</template>

<style></style>

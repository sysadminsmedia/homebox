<script setup lang="ts">
  import type { TreeItem } from "~~/lib/api/types/data-contracts";
  import LocationTreeNode from "./Node.vue";

  type Props = {
    locs: TreeItem[];
    treeId: string;
  };

  const props = defineProps<Props>();

  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });

  const sortedLocs = computed(() => {
    const list = props.locs ?? [];
    return [...list].sort((a, b) => collator.compare(a.name, b.name));
  });
</script>

<template>
  <div>
    <p v-if="sortedLocs.length === 0" class="text-center text-sm">
      {{ $t("location.tree.no_locations") }}
    </p>
    <LocationTreeNode v-for="item in sortedLocs" :key="item.id" :item="item" :tree-id="treeId" />
  </div>
</template>

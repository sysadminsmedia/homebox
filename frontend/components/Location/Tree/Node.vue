<script setup lang="ts">
  import { useTreeState } from "./tree-state";
  import type { TreeItem } from "~~/lib/api/types/data-contracts";
  import MdiChevronRight from "~icons/mdi/chevron-right";
  import MdiMapMarker from "~icons/mdi/map-marker";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import LocationTreeNode from "./Node.vue";

  type Props = {
    treeId: string;
    item: TreeItem;
    showItems?: boolean;
  };
  const props = withDefaults(defineProps<Props>(), {
    showItems: true,
  });

  const link = computed(() => {
    return props.item.type === "location" ? `/location/${props.item.id}` : `/item/${props.item.id}`;
  });

  const state = useTreeState(props.treeId);

  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });

  const filteredChildren = computed(() => {
    const children = props.item.children ?? [];

    if (props.showItems) {
      return children;
    }

    return children.filter(child => child.type === "location");
  });

  const sortedChildren = computed(() => {
    return [...filteredChildren.value].sort((a, b) => collator.compare(a.name, b.name));
  });

  const hasChildren = computed(() => filteredChildren.value.length > 0);

  const openRef = computed({
    get() {
      return state.value[nodeHash.value] ?? false;
    },
    set(value: boolean) {
      state.value[nodeHash.value] = value;
    },
  });

  const nodeHash = computed(() => {
    // converts a UUID to a short hash
    return props.item.id.replace(/-/g, "").substring(0, 8);
  });
</script>

<template>
  <div>
    <div
      class="flex items-center gap-1 rounded p-1"
      :class="{
        'cursor-pointer hover:bg-accent hover:text-accent-foreground': hasChildren,
      }"
      @click="openRef = !openRef"
    >
      <div
        class="mr-1 flex items-center justify-center rounded p-0.5"
        :class="{
          'hover:bg-accent hover:text-accent-foreground': hasChildren,
        }"
      >
        <div v-if="!hasChildren" class="size-6" />
        <div v-else class="group/node relative size-6" :data-swap="openRef">
          <div
            class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-data-[swap=true]/node:rotate-90"
          >
            <MdiChevronRight class="size-6" />
          </div>
        </div>
      </div>
      <MdiMapMarker v-if="item.type === 'location'" class="size-4" />
      <MdiPackageVariant v-else class="size-4" />
      <NuxtLink class="text-lg hover:underline" :to="link" @click.stop>{{ item.name }} </NuxtLink>
    </div>
    <div v-if="openRef" class="ml-4">
      <LocationTreeNode
        v-for="child in sortedChildren"
        :key="child.id"
        :item="child"
        :tree-id="treeId"
        :show-items="showItems"
      />
    </div>
  </div>
</template>

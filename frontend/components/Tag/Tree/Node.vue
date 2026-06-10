<script setup lang="ts">
  import MdiChevronRight from "~icons/mdi/chevron-right";
  import { getContrastTextColor } from "~/lib/utils";
  import { getIconComponent } from "~/lib/icons";
  import { useTreeState } from "./tree-state";
  import type { TagTreeItem } from "./types";
  import TagTreeNode from "./Node.vue";

  type Props = {
    treeId: string;
    item: TagTreeItem;
  };

  const props = defineProps<Props>();

  const tagIcon = computed(() => getIconComponent(props.item.icon));

  const state = useTreeState(props.treeId);

  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });

  const sortedChildren = computed(() => {
    return [...(props.item.children ?? [])].sort((a, b) => collator.compare(a.name, b.name));
  });

  const hasChildren = computed(() => sortedChildren.value.length > 0);

  const nodeHash = computed(() => {
    return props.item.id.replace(/-/g, "").substring(0, 8);
  });

  const openRef = computed({
    get() {
      return state.value[nodeHash.value] ?? false;
    },
    set(value: boolean) {
      state.value[nodeHash.value] = value;
    },
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

      <div
        class="mr-1 flex size-5 items-center justify-center rounded-full"
        :style="
          item.color
            ? { backgroundColor: item.color, color: getContrastTextColor(item.color) }
            : { backgroundColor: 'hsl(var(--accent))', color: 'hsl(var(--accent-foreground))' }
        "
      >
        <component :is="tagIcon" class="size-3" />
      </div>

      <NuxtLink class="text-lg hover:underline" :to="`/tag/${item.id}`" @click.stop>{{ item.name }}</NuxtLink>
    </div>

    <div v-if="openRef" class="ml-4">
      <TagTreeNode v-for="child in sortedChildren" :key="child.id" :item="child" :tree-id="treeId" />
    </div>
  </div>
</template>

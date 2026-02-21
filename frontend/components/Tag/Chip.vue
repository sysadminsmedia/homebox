<script setup lang="ts">
  import type { TagOut, TagSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiTagOutline from "~icons/mdi/tag-outline";
  import { getContrastTextColor } from "~/lib/utils";

  export type sizes = "sm" | "md" | "lg" | "xl";
  defineProps({
    tag: {
      type: Object as () => TagOut | TagSummary,
      required: true,
    },
    size: {
      type: String as () => sizes,
      default: "md",
    },
    hideIcon: {
      type: Boolean,
      default: false,
    },
  });
</script>

<template>
  <NuxtLink
    class="group/tag-chip flex gap-2 rounded-full shadow transition duration-300 hover:bg-accent/50"
    :class="{
      'p-4 py-1 text-base': size === 'lg',
      'p-3 py-1 text-sm': size !== 'sm' && size !== 'lg',
      'p-2 py-0.5 text-xs': size === 'sm',
    }"
    :style="
      tag.color
        ? { backgroundColor: tag.color, color: getContrastTextColor(tag.color) }
        : { backgroundColor: 'hsl(var(--accent))' }
    "
    :to="`/tag/${tag.id}`"
  >
    <template v-if="!hideIcon">
      <div class="relative">
        <MdiTagOutline class="invisible" /><!-- hack to ensure the size is correct -->

        <div
          class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover/tag-chip:rotate-90"
        >
          <MdiTagOutline class="group-hover/tag-chip:hidden" />
          <MdiArrowUp class="hidden group-hover/tag-chip:block" />
        </div>
      </div>
    </template>
    {{ tag.name }}
  </NuxtLink>
</template>

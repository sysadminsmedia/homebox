<script setup lang="ts">
  import type { LabelOut, LabelSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiTagOutline from "~icons/mdi/tag-outline";
  import { getContrastTextColor } from "~/lib/utils";

  export type sizes = "sm" | "md" | "lg" | "xl";
  defineProps({
    label: {
      type: Object as () => LabelOut | LabelSummary,
      required: true,
    },
    size: {
      type: String as () => sizes,
      default: "md",
    },
  });
</script>

<template>
  <NuxtLink
    class="group/label-chip flex gap-2 rounded-full shadow transition duration-300 hover:bg-accent/50"
    :class="{
      'p-4 py-1 text-base': size === 'lg',
      'p-3 py-1 text-sm': size !== 'sm' && size !== 'lg',
      'p-2 py-0.5 text-xs': size === 'sm',
    }"
    :style="
      label.color
        ? { backgroundColor: label.color, color: getContrastTextColor(label.color) }
        : { backgroundColor: 'hsl(var(--accent))' }
    "
    :to="`/label/${label.id}`"
  >
    <div class="relative">
      <MdiTagOutline class="invisible" /><!-- hack to ensure the size is correct -->

      <div
        class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover/label-chip:rotate-90"
      >
        <MdiTagOutline class="group-hover/label-chip:hidden" />
        <MdiArrowUp class="hidden group-hover/label-chip:block" />
      </div>
    </div>
    {{ label.name }}
  </NuxtLink>
</template>

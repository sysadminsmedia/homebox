<script setup lang="ts">
  import type { TagOut, TagSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import { getContrastTextColor } from "~/lib/utils";
  import { getIconComponent } from "~/lib/icons";

  export type sizes = "sm" | "md" | "lg" | "xl";
  const props = defineProps({
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
    ancestors: {
      type: Boolean,
      default: false,
    },
  });

  const chipIcon = computed(() => {
    return getIconComponent(props.tag.icon);
  });

  // Custom colors are arbitrary, so hover tints them with an overlay instead of a Tailwind
  // color class: darken when the text is white, lighten when it is black, so contrast improves.
  const customColorStyle = computed(() => {
    if (!props.tag.color) {
      return undefined;
    }
    const textColor = getContrastTextColor(props.tag.color);
    return {
      backgroundColor: props.tag.color,
      color: textColor,
      "--chip-hover-overlay": textColor === "#fff" ? "rgb(0 0 0 / 0.15)" : "rgb(255 255 255 / 0.3)",
    };
  });
</script>

<template>
  <NuxtLink
    class="group/tag-chip flex gap-2 rounded-full border shadow transition duration-300"
    :class="{
      'p-4 py-1 text-base': size === 'lg',
      'p-3 py-1 text-sm': size !== 'sm' && size !== 'lg',
      'p-2 py-0.5 text-xs': size === 'sm',
      'border-dashed italic': ancestors,
      'border-transparent': !ancestors,
      'bg-accent text-accent-foreground hover:bg-accent/80': !tag.color,
      'hover:bg-[linear-gradient(var(--chip-hover-overlay),var(--chip-hover-overlay))]': tag.color,
    }"
    :style="customColorStyle"
    :to="`/tag/${tag.id}`"
  >
    <template v-if="!hideIcon">
      <div class="relative">
        <component :is="chipIcon" class="invisible" /><!-- hack to ensure the size is correct -->

        <div
          class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover/tag-chip:rotate-90"
        >
          <component :is="chipIcon" class="group-hover/tag-chip:hidden" />
          <MdiArrowUp class="hidden group-hover/tag-chip:block" />
        </div>
      </div>
    </template>
    {{ tag.name }}
  </NuxtLink>
</template>

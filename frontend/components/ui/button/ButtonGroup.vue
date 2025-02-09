<script setup lang="ts">
  import { type HTMLAttributes } from "vue";
  import { cn } from "@/lib/utils";

  const props = defineProps<{ class?: HTMLAttributes["class"] }>();
</script>

<template>
  <div
    :class="
      cn(
        'inline-flex rounded-lg',
        '[&>*]:rounded-none',
        '[&>*:first-child]:rounded-l-lg',
        '[&>*:last-child]:rounded-r-lg',
        props.class
      )
    "
  >
    <slot />
  </div>
</template>

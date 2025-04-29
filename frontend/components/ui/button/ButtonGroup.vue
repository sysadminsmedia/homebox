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
        '[&>[data-button=true]]:rounded-none',
        '[&>[data-button=true]:first-child]:rounded-l-lg',
        '[&>[data-button=true]:last-child]:rounded-r-lg',
        '[&_[data-button=true]]:rounded-none',
        '[&_[data-button=true][data-pos=start]]:rounded-l-lg',
        '[&_[data-button=true][data-pos=end]]:rounded-r-lg',
        props.class
      )
    "
  >
    <slot />
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

const props = defineProps<{
  keys: string[];
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}>();

const sizeClasses = computed(() => {
  return {
    sm: 'px-1 py-0.5 text-xs',
    md: 'px-1.5 py-0.5 text-sm',
    lg: 'px-2 py-1 text-base',
  }[props.size ?? 'md'];
});
</script>

<template>
  <kbd
    :class="[
      'inline-flex items-center gap-0.5 rounded-md bg-background shadow-sm ring-1 ring-gray-200',
      sizeClasses,
      className,
    ]"
  >
    <span
      v-for="(key, index) in keys"
      :key="index"
      class="font-medium text-muted-foreground font-sans"
    >
      {{ key }}
    </span>
  </kbd>
</template>

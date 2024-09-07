<script lang="ts" setup>
  defineProps({
    cmp: {
      type: String,
      default: "div",
    },
  });
</script>

<template>
  <component :is="cmp" class="container mx-auto max-w-6xl px-3">
    <slot />
  </component>
</template>

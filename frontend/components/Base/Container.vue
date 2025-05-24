<script lang="ts" setup>
  defineProps({
    cmp: {
      type: String,
      default: "div",
    },
  });
</script>

<template>
  <component :is="cmp" class="container mx-auto my-4 max-w-7xl px-4">
    <slot />
  </component>
</template>

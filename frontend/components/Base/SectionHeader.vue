<template>
  <div class="pb-3">
    <h3
      class="flex items-center text-3xl font-bold tracking-tight"
      :class="{
        'text-neutral-content': dark,
        'text-content': !dark,
      }"
    >
      <slot />
    </h3>
    <p v-if="$slots.description" class="mt-2 max-w-4xl text-sm text-base-content">
      <slot name="description" />
    </p>
    <div v-if="$slots.after">
      <slot name="after" />
    </div>
  </div>
</template>

<script lang="ts" setup>
  defineProps({
    dark: {
      type: Boolean,
      default: false,
    },
  });
</script>

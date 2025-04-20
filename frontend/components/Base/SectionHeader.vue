<template>
  <div class="pb-3">
    <CardTitle
      class="flex items-center"
      :class="{
        'text-neutral-content': dark,
      }"
    >
      <slot />
    </CardTitle>
    <CardDescription v-if="$slots.description">
      <slot name="description" />
    </CardDescription>
    <div v-if="$slots.after">
      <slot name="after" />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { CardDescription, CardTitle } from "@/components/ui/card";

  defineProps({
    dark: {
      type: Boolean,
      default: false,
    },
  });
</script>

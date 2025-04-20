<template>
  <div class="grid grid-cols-1 gap-10 py-6 md:grid-cols-4">
    <div class="col-span-3">
      <h4 class="mb-1 text-lg font-semibold">
        <slot name="title"></slot>
      </h4>
      <p class="text-sm">
        <slot></slot>
      </p>
    </div>
    <template v-if="to">
      <NuxtLink class="mt-auto" :to="to" :class="buttonVariants({ size: 'lg' })">
        <slot name="button">
          <slot name="title"></slot>
        </slot>
      </NuxtLink>
    </template>
    <template v-else>
      <Button class="mt-auto" size="lg" @click="$emit('action')">
        <slot name="button">
          <slot name="title"></slot>
        </slot>
      </Button>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { defineProps } from "vue";
  import { Button, buttonVariants } from "@/components/ui/button";

  defineProps<{
    to?: string;
  }>();

  defineEmits(["action"]);
</script>

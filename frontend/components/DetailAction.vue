<template>
  <div class="flex flex-col gap-10 py-6 md:flex-row">
    <div class="flex-1">
      <h4 class="mb-1 text-lg font-semibold">
        <slot name="title" />
      </h4>
      <p class="text-sm">
        <slot />
      </p>
    </div>
    <div class="flex items-center">
      <template v-if="to">
        <NuxtLink :to="to" :class="buttonVariants({ size: 'lg' })" class="min-w-52 grow">
          <slot name="button">
            <slot name="title" />
          </slot>
        </NuxtLink>
      </template>
      <template v-else>
        <Button class="min-w-52 grow" size="lg" @click="$emit('action')">
          <slot name="button">
            <slot name="title" />
          </slot>
        </Button>
      </template>
    </div>
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

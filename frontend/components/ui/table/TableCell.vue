<script setup lang="ts">
  import type { HTMLAttributes } from "vue";
  import { cn } from "@/lib/utils";

  const props = defineProps<{
    class?: HTMLAttributes["class"];
    href?: string;
    compact?: boolean;
  }>();
</script>

<template>
  <td :class="cn('p-0 align-middle relative', props.class)">
    <NuxtLink
      v-if="props.href"
      :to="props.href"
      class="block size-full"
      :class="props.compact ? 'p-0' : 'p-4'"
      style="min-width: 100%; min-height: 100%; display: flex; align-items: stretch"
    >
      <span class="flex min-h-0 min-w-0 flex-1 items-center"> <slot /> </span>
    </NuxtLink>
    <template v-else>
      <div
        :class="
          cn(
            'flex size-full min-h-0 min-w-0 items-center [&:has([role=checkbox])]:pr-0',
            props.compact ? 'justify-center p-0' : 'p-4'
          )
        "
      >
        <slot />
      </div>
    </template>
  </td>
</template>

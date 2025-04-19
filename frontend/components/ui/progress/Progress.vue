<script setup lang="ts">
  import { ProgressIndicator, ProgressRoot, type ProgressRootProps } from "reka-ui";
  import { computed, type HTMLAttributes } from "vue";
  import { cn } from "@/lib/utils";

  const props = withDefaults(defineProps<ProgressRootProps & { class?: HTMLAttributes["class"] }>(), {
    modelValue: 0,
  });

  const value = computed(() => {
    return props.modelValue ?? 0;
  });

  const delegatedProps = computed(() => {
    const { class: _, ...delegated } = props;

    return delegated;
  });
</script>

<template>
  <ProgressRoot
    v-bind="delegatedProps"
    :class="cn('relative h-2 w-full overflow-hidden rounded-full bg-secondary', props.class)"
  >
    <ProgressIndicator
      class="size-full flex-1 transition-all"
      :style="`transform: translateX(-${100 - value}%);`"
      :class="{
        'bg-green-500': value > 50,
        'bg-yellow-500': value > 25 && value < 50,
        'bg-red-500': value < 25,
      }"
    />
  </ProgressRoot>
</template>

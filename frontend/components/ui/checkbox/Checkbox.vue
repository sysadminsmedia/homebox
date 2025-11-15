<script setup lang="ts">
  import type { CheckboxRootEmits, CheckboxRootProps } from "reka-ui";
  import { cn } from "@/lib/utils";
  import { Check, Minus } from "lucide-vue-next";
  import { CheckboxIndicator, CheckboxRoot, useForwardPropsEmits } from "reka-ui";
  import { computed, type HTMLAttributes } from "vue";

  const props = defineProps<CheckboxRootProps & { class?: HTMLAttributes["class"] }>();
  const emits = defineEmits<CheckboxRootEmits>();

  const delegatedProps = computed(() => {
    const { class: _, ...delegated } = props;

    return delegated;
  });

  const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <CheckboxRoot
    v-bind="forwarded"
    :class="
      cn(
        'peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground data-[state=indeterminate]:bg-primary data-[state=indeterminate]:text-primary-foreground',
        props.class
      )
    "
  >
    <CheckboxIndicator class="flex size-full items-center justify-center text-current">
      <slot>
        <Check v-if="typeof props.modelValue === 'boolean'" class="size-4" />
        <Minus v-else class="size-4" />
      </slot>
    </CheckboxIndicator>
  </CheckboxRoot>
</template>

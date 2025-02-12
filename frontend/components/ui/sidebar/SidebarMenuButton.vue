<script setup lang="ts">
  import { type Component, computed } from "vue";
  import SidebarMenuButtonChild, { type SidebarMenuButtonProps } from "./SidebarMenuButtonChild.vue";
  import { useSidebar } from "./utils";
  import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

  defineOptions({
    inheritAttrs: false,
  });

  const props = withDefaults(
    defineProps<
      SidebarMenuButtonProps & {
        tooltip?: string | Component;
        hotkey?: string;
      }
    >(),
    {
      as: "button",
      variant: "default",
      size: "default",
    }
  );

  const { isMobile, state } = useSidebar();

  const delegatedProps = computed(() => {
    const { tooltip, hotkey, ...delegated } = props;
    return delegated;
  });
</script>

<template>
  <SidebarMenuButtonChild v-if="!tooltip" v-bind="{ ...delegatedProps, ...$attrs }">
    <slot />
  </SidebarMenuButtonChild>

  <Tooltip v-else>
    <TooltipTrigger as-child>
      <SidebarMenuButtonChild v-bind="{ ...delegatedProps, ...$attrs }" :size="state === 'collapsed' ? 'default' : 'lg'" :class="state === 'collapsed' ? '' : 'text-xl'">
        <slot />
      </SidebarMenuButtonChild>
    </TooltipTrigger>
    <TooltipContent side="right" align="center" :hidden="state !== 'collapsed' || isMobile">
      <template v-if="typeof tooltip === 'string'">
        {{ tooltip }}
      </template>
      <component :is="tooltip" v-else />
    </TooltipContent>
    <TooltipContent v-if="hotkey" :hidden="isMobile">
      {{ hotkey }}
    </TooltipContent>
  </Tooltip>
</template>

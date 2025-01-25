<script setup lang="ts">
  import type { HTMLAttributes } from "vue";
  import { useSidebar } from "./utils";
  import { Button } from "@/components/ui/button";
  import { cn } from "@/lib/utils";
  import MdiMenu from "~icons/mdi/menu";

  const props = defineProps<{
    class?: HTMLAttributes["class"];
    variant?: "ghost" | "default";
  }>();

  const { toggleSidebar } = useSidebar();
</script>

<template>
  <Button data-sidebar="trigger" :variant="props.variant ?? 'ghost'" size="icon" :class="cn('h-7 w-7', props.class)" @click="toggleSidebar">
    <MdiMenu class="text-primary-foreground" />
    <span class="sr-only">Toggle Sidebar</span>
  </Button>
</template>

<script setup lang="ts">
  import { useSidebar } from "./utils";
  import SidebarMenuButton from "./SidebarMenuButton.vue";

  const { setOpenMobile } = useSidebar();

  const props = defineProps<{
    href: string;
  }>();
</script>

<template>
  <SidebarMenuButton as-child>
    <NuxtLink :to="props.href" @click="setOpenMobile(false)">
      <slot />
    </NuxtLink>
  </SidebarMenuButton>
</template>

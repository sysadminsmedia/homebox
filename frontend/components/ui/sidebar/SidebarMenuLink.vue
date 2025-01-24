<script setup lang="ts">
  import { useRouter } from "vue-router";
  import { useSidebar } from "./utils";
  import SidebarMenuButton from "./SidebarMenuButton.vue";

  const { setOpenMobile } = useSidebar();

  const props = defineProps<{
    href: string;
  }>();
</script>

<template>
  <SidebarMenuButton as-child>
    <NuxtLink :to="props.href" @click.prevent="setOpenMobile(false)">
      <slot />
    </NuxtLink>
  </SidebarMenuButton>
</template>

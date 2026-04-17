<script setup lang="ts">
  import BaseContainer from "@/components/Base/Container.vue";
  definePageMeta({
    middleware: ["auth"],
  });
</script>

<template>
  <BaseContainer>
    <NuxtPage />
  </BaseContainer>
</template>

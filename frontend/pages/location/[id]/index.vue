<script setup lang="ts"></script>

<template>
  <BaseContainer>
    <NuxtPage />
  </BaseContainer>
</template>

<script lang="ts">
  import BaseContainer from "@/components/Base/Container.vue";
  definePageMeta({
    middleware: ["auth"],
  });
</script>

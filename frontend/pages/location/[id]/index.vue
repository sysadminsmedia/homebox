<script setup lang="ts">
  definePageMeta({
    middleware: ["auth"],
  });
</script>

<template>
  <BaseContainer>
    <NuxtPage />
  </BaseContainer>
</template>

<script lang="ts">
  import BaseContainer from "@/components/Base/Container.vue";
</script>

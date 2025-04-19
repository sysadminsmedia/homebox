<script setup lang="ts">
  import { buttonVariants } from "@/components/ui/button";

  useHead({
    title: "404. Not Found",
  });
  definePageMeta({
    layout: "404",
  });
</script>

<template>
  <h1 class="flex flex-col gap-5 text-center font-extrabold">
    <span class="text-7xl">404.</span>
    <span class="text-5xl"> Page Not Found </span>
    <NuxtLink to="/" :class="buttonVariants({ size: 'lg' })"> Return Home </NuxtLink>
  </h1>
</template>

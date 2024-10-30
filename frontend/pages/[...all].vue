<script setup lang="ts">
  useHead({
    title: "404. Not Found",
  });
  definePageMeta({
    layout: "404",
  });
</script>

<template>
  <h1 class="flex flex-col text-center font-extrabold">
    <span class="text-7xl">404.</span>
    <span class="mt-5 text-5xl"> Page Not Found </span>
    <NuxtLink to="/" class="btn mt-5 text-xl"> Return Home </NuxtLink>
  </h1>
</template>

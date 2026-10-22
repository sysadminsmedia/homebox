<script setup lang="ts">
  // Printers moved under Collection; keep the old URL working for bookmarks and links.
  definePageMeta({
    middleware: ["auth"],
  });

  await navigateTo("/collection/printers", { replace: true });
</script>

<template>
  <div />
</template>

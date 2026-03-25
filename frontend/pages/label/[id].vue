<script setup lang="ts">
  // Backwards compatibility redirect from /label/:id to /tag/:id
  const route = useRoute();
  const tagId = route.params.id as string;

  // Redirect to the new tag route
  await navigateTo(`/tag/${tagId}`, { replace: true, redirectCode: 308 });
</script>

<template>
  <div></div>
</template>

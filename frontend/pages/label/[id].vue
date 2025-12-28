<script setup lang="ts">
  // Backwards compatibility redirect from /label/:id to /tag/:id
  const route = useRoute();
  const labelId = route.params.id as string;

  // Redirect to the new tag route
  await navigateTo(`/tag/${labelId}`, { replace: true, redirectCode: 301 });
</script>

<template>
  <div>
    <!-- Redirecting... -->
  </div>
</template>

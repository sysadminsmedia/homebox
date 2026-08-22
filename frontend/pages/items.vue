<script setup lang="ts">
  // Backwards compatibility redirect from /items to /search
  const route = useRoute();

  console.log(route);

  // Redirect to the new tag route
  // await navigateTo(`/search`, { replace: true, redirectCode: 308, query });
  await navigateTo(
    {
      path: "/search",
      query: route.query,
    },
    {
      redirectCode: 308,
      replace: true,
    }
  );
</script>

<template>
  <div></div>
</template>

<script setup lang="ts">
  definePageMeta({
    middleware: ["auth"],
  });

  if (!ctx.isAuthorized()) {
      toast.error("User is not logged in");
      navigateTo("/lostandfound");
      return;
    }

  const route = useRoute();
  const assetId = computed<string>(() => route.params.id as string);
  await navigateTo("/assets/" + assetId.value, { replace: true, redirectCode: 301 });
</script>

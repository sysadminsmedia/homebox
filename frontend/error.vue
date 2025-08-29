<script setup lang="ts">
  import type { NuxtError } from "#app";
  import { NuxtLink } from "#components";
  import { buttonVariants } from "@/components/ui/button";

  const props = defineProps({
    error: {
      type: Object as () => NuxtError,
      required: true,
    },
  });

  console.log(props.error);
</script>

<template>
  <main class="grid min-h-screen w-full place-items-center">
    <h1 class="flex flex-col gap-5 text-center font-extrabold">
      <span class="text-7xl">{{ error.statusCode }}.</span>
      <span class="text-5xl"> {{ error.message }} </span>
      <NuxtLink to="/" :class="buttonVariants({ size: 'lg' })"> {{ $t("global.return_home") }} </NuxtLink>
    </h1>
  </main>
</template>

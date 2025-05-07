<script setup lang="ts">
  import type { NuxtError } from "#app";
  import { buttonVariants } from "@/components/ui/button";

  const props = defineProps({
    // eslint-disable-next-line vue/require-default-prop
    error: Object as () => NuxtError,
  });

  console.log(props.error);
</script>

<template>
  <main class="grid min-h-screen w-full place-items-center">
    <h1 class="flex flex-col gap-5 text-center font-extrabold">
      <span class="text-7xl">{{ error?.statusCode }}.</span>
      <span class="text-5xl"> {{ error?.message }} </span>
      <NuxtLink to="/" :class="buttonVariants({ size: 'lg' })"> Return Home </NuxtLink>
    </h1>
  </main>
</template>

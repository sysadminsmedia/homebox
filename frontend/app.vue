<template>
  <DialogProvider>
    <ClientOnly>
      <Toaster class="pointer-events-auto" />
    </ClientOnly>

    <NuxtLayout>
      <Html :lang="locale" :data-theme="theme || 'homebox'" />
      <Link rel="icon" type="image/svg" href="/favicon.svg"></Link>
      <Link rel="apple-touch-icon" href="/apple-touch-icon.png" size="180x180" />
      <Link rel="mask-icon" href="/mask-icon.svg" color="#5b7f67" />
      <Meta name="theme-color" content="#5b7f67" />
      <Link rel="manifest" href="/manifest.webmanifest" />
      <NuxtPage />
    </NuxtLayout>
  </DialogProvider>
</template>

<script lang="ts" setup>
  import { useI18n } from "vue-i18n";
  import { DialogProvider } from "@/components/ui/dialog-provider";
  import { Toaster } from "@/components/ui/sonner";

  const { theme } = useTheme();

  const { locale } = useI18n();
</script>

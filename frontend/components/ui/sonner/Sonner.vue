<script lang="ts" setup>
import { Toaster as Sonner, type ToasterProps } from 'vue-sonner'

const props = defineProps<ToasterProps>()
</script>

<template>
  <!-- TODO: once daisyui is removed, remove shadcn- prefix -->
  <Sonner
    class="shadcn-toaster group"
    v-bind="props"
    :toast-options="{
      classes: {
        toast: 'group shadcn-toast group-[.shadcn-toaster]:bg-background group-[.shadcn-toaster]:text-foreground group-[.shadcn-toaster]:border-border group-[.shadcn-toaster]:shadow-lg',
        description: 'group-[.shadcn-toast]:text-muted-foreground',
        actionButton:
          'group-[.shadcn-toast]:bg-primary group-[.shadcn-toast]:text-primary-foreground',
        cancelButton:
          'group-[.shadcn-toast]:bg-muted group-[.shadcn-toast]:text-muted-foreground',
      },
    }"
  />
</template>

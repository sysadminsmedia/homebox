<script setup lang="ts">
  import { toast } from "vue-sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import { Shortcut } from "~/components/ui/shortcut";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
</script>

<template>
  <h1 class="text-3xl">Testing</h1>
  <p>This is a page for testing during the transition to shadcn.</p>
  <Button
    @click="
      toast.success('This is a test', {
        duration: 100000,
      })
    "
    >Create Toast</Button
  >
  <ButtonGroup>
    <Button @click="toast.success('Success')">Success</Button>
    <Button variant="outline" @click="toast.info('Info')">Info</Button>
    <Button variant="destructive" @click="toast.warning('Warning')">Warning</Button>
    <Button variant="destructive" @click="toast.error('Error')">Error</Button>
    <Button>Button 3</Button>
  </ButtonGroup>
  <span class="flex items-center gap-1 text-sm">
    Use
    <Shortcut size="sm" :keys="['Shift']" /> + <Shortcut size="sm" :keys="['Enter']" /> to create and add another.
  </span>
  <div class="group relative size-12">
    <div
      class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover:rotate-180"
    >
      <MdiPackageVariant class="size-5 group-hover:hidden" />
      <MdiPackageVariantClosed class="hidden size-5 group-hover:block" />
    </div>
  </div>
</template>

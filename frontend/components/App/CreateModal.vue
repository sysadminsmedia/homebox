<template>
  <Dialog v-if="isDesktop" :dialog-id="dialogId">
    <DialogScrollContent>
      <DialogHeader>
        <DialogTitle>{{ title }}</DialogTitle>
      </DialogHeader>

      <slot />

      <DialogFooter>
        <span class="flex items-center gap-1 text-sm">
          Use <Shortcut size="sm" :keys="['Shift']" /> + <Shortcut size="sm" :keys="['Enter']" /> to create and add
          another.
        </span>
      </DialogFooter>
    </DialogScrollContent>
  </Dialog>

  <Drawer v-else :dialog-id="dialogId">
    <DrawerContent class="max-h-[90%]">
      <DrawerHeader>
        <DrawerTitle>{{ title }}</DrawerTitle>
      </DrawerHeader>

      <div class="m-2 overflow-y-auto p-2">
        <slot />
      </div>
    </DrawerContent>
  </Drawer>
</template>

<script setup lang="ts">
  import { useMediaQuery } from "@vueuse/core";
  import { Drawer, DrawerContent, DrawerHeader, DrawerTitle } from "@/components/ui/drawer";
  import { Dialog, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";

  const isDesktop = useMediaQuery("(min-width: 768px)");

  defineProps<{
    dialogId: string;
    title: string;
  }>();
</script>

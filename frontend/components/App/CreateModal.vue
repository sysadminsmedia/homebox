<template>
  <Dialog v-if="isDesktop" :dialog-id="dialogId">
    <DialogScrollContent>
      <DialogHeader>
        <div class="mr-4 flex place-items-center justify-between">
          <DialogTitle>{{ title }}</DialogTitle>
          <slot name="header-actions" />
        </div>
      </DialogHeader>

      <slot />

      <DialogFooter>
        <i18n-t
          keypath="components.app.create_modal.createAndAddAnother"
          tag="span"
          class="flex items-center gap-1 text-sm"
        >
          <template #shiftKey>
            <Shortcut size="sm" :keys="[$t('components.app.create_modal.shift')]" />
          </template>
          <template #enterKey>
            <Shortcut size="sm" :keys="[$t('components.app.create_modal.enter')]" />
          </template>
        </i18n-t>
      </DialogFooter>
    </DialogScrollContent>
  </Dialog>

  <Drawer v-else :dialog-id="dialogId">
    <DrawerContent class="max-h-[90%]">
      <DrawerHeader>
        <DrawerTitle>{{ title }}</DrawerTitle>
      </DrawerHeader>
      <div class="flex justify-center">
        <slot name="header-actions" />
      </div>

      <div class="m-2 overflow-y-auto p-2">
        <slot />
      </div>
    </DrawerContent>
  </Drawer>
</template>

<script setup lang="ts">
  import { useMediaQuery } from "@vueuse/core";
  import type { DialogID } from "@/components/ui/dialog-provider/utils";
  import { Drawer, DrawerContent, DrawerHeader, DrawerTitle } from "@/components/ui/drawer";
  import { Dialog, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";

  const isDesktop = useMediaQuery("(min-width: 768px)");

  defineProps<{
    dialogId: DialogID;
    title: string;
  }>();
</script>

<!-- DialogProvider.vue -->
<script setup lang="ts">
  import { ref, computed } from "vue";
  // import { useEventListener } from "@vueuse/core";
  import { provideDialogContext } from "./utils";

  const activeDialog = ref<string | null>(null);

  const openDialog = (dialogId: string) => {
    activeDialog.value = dialogId;
  };

  const closeDialog = (dialogId?: string) => {
    if (dialogId) {
      if (activeDialog.value === dialogId) {
        activeDialog.value = null;
      }
    } else {
      activeDialog.value = null;
    }
  };

  // Provide context to child components
  provideDialogContext({
    activeDialog: computed(() => activeDialog.value),
    openDialog,
    closeDialog,
  });

  // Optionally, listen to keyboard events for dialog toggles (for example, use the 'd' key)
  // useEventListener("keydown", (event: KeyboardEvent) => {
  //   if (event.key === "d" && (event.metaKey || event.ctrlKey)) {
  //     event.preventDefault();
  //     toggleDialog("dialog1"); // Example: toggle 'dialog1' when 'Ctrl/Cmd + D' is pressed
  //   }
  // });
</script>

<template>
  <slot />
</template>

<!-- DialogProvider.vue -->
<script setup lang="ts">
  import { ref, reactive, computed } from "vue";
  import { provideDialogContext } from "./utils";

  const activeDialog = ref<string | null>(null);
  const activeAlerts = reactive<string[]>([]);

  const openDialog = (dialogId: string) => {
    if (activeAlerts.length > 0) return;
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

  const addAlert = (alertId: string) => {
    activeAlerts.push(alertId);
  };

  const removeAlert = (alertId: string) => {
    const index = activeAlerts.indexOf(alertId);
    if (index !== -1) {
      activeAlerts.splice(index, 1);
    }
  };

  // Provide context to child components
  provideDialogContext({
    activeDialog: computed(() => activeDialog.value),
    openDialog,
    closeDialog,
    activeAlerts: computed(() => activeAlerts),
    addAlert,
    removeAlert,
  });
</script>

<template>
  <slot />
</template>

<!-- DialogProvider.vue -->
<script setup lang="ts">
  import { ref, reactive, computed } from "vue";
  import { provideDialogContext, ActiveDialog } from "./utils";

  const activeDialog =  ref<ActiveDialog | null>(null);
  const activeAlerts = reactive<string[]>([]);

  const openDialog = (dialogId: string, params?: any) => {
    if (activeAlerts.length > 0) return;

    const ad = new ActiveDialog(dialogId, params);
    activeDialog.value = ad;
  };

  const closeDialog = (dialogId?: string) => {
    if (dialogId) {
      if (activeDialog.value && activeDialog.value.id === dialogId) {
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

<!-- DialogProvider.vue -->
<script setup lang="ts">
  import { ref, reactive, computed } from "vue";
  import { provideDialogContext, type DialogID, type DialogParamsMap } from "./utils";

  const activeDialog =  ref<DialogID | null>(null);
  const activeAlerts = reactive<string[]>([]);
  const openDialogCallbacks = new Map<DialogID, (params: any) => void>();

  const registerOpenDialogCallback = <T extends DialogID>(
    dialogId: T,
    callback: (params?: T extends keyof DialogParamsMap ? DialogParamsMap[T] : undefined) => void
  ) =>
  {
    openDialogCallbacks.set(dialogId, callback as (params: any) => void);
  }

  const openDialog = (dialogId: DialogID, params?: any) => {
    if (activeAlerts.length > 0) return;

    activeDialog.value = dialogId;

    const openCallback = openDialogCallbacks.get(dialogId);
    if (openCallback) {
      openCallback(params);
    }
  };

  const closeDialog = (dialogId?: DialogID) => {
    if (dialogId) {
      if (activeDialog.value && activeDialog.value === dialogId) {
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
    registerOpenDialogCallback,
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

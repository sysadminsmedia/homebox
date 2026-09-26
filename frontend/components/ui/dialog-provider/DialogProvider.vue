<script setup lang="ts">
  import { ref, reactive, computed } from 'vue';
  import {
    provideDialogContext,
    type DialogID,
    type DialogParamsMap,
  } from './utils';

  const activeDialog = ref<DialogID | null>(null);
  const activeAlerts = reactive<string[]>([]);
  /** Multiple components may register for the same dialog (e.g. two MaintenanceEditModal trees); invoke all. */
  const openDialogCallbacks = new Map<DialogID, Set<(params: any) => void>>();

  // onClose for the currently-open dialog (only one dialog can be active)
  let activeOnCloseCallback: ((result?: any) => void) | undefined;

  const registerOpenDialogCallback = <T extends DialogID>(
    dialogId: T,
    callback: (params?: T extends keyof DialogParamsMap ? DialogParamsMap[T] : undefined) => void
  ) => {
    const cb = callback as (params: any) => void;
    let set = openDialogCallbacks.get(dialogId);
    if (!set) {
      set = new Set();
      openDialogCallbacks.set(dialogId, set);
    }
    set.add(cb);
    return () => {
      const s = openDialogCallbacks.get(dialogId);
      if (!s) {
        return;
      }
      s.delete(cb);
      if (s.size === 0) {
        openDialogCallbacks.delete(dialogId);
      }
    };
  };

  const openDialog = <T extends DialogID>(dialogId: T, options?: any) => {
    if (activeAlerts.length > 0) return;

    activeDialog.value = dialogId;
    activeOnCloseCallback = options?.onClose;

    const callbacks = openDialogCallbacks.get(dialogId);
    if (callbacks) {
      for (const openCallback of callbacks) {
        openCallback(options?.params);
      }
    }
  };

  function closeDialog(dialogId?: DialogID, result?: any) {
    // No dialogId passed -> close current active dialog without result
    if (!dialogId) {
      if (activeDialog.value) {
        // call onClose (if any) with no result
        activeOnCloseCallback?.(undefined);
        activeOnCloseCallback = undefined;
      }
      activeDialog.value = null;
      return;
    }

    // dialogId passed -> if it's the active dialog, call onClose with result
    if (activeDialog.value && activeDialog.value === dialogId) {
      activeOnCloseCallback?.(result);
      activeOnCloseCallback = undefined;
      activeDialog.value = null;
    }
  }

  const addAlert = (alertId: string) => {
    activeAlerts.push(alertId);
  };

  const removeAlert = (alertId: string) => {
    const index = activeAlerts.indexOf(alertId);
    if (index !== -1) activeAlerts.splice(index, 1);
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

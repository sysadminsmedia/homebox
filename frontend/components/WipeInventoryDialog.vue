<template>
  <AlertDialog :open="dialog" @update:open="handleOpenChange">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ $t("tools.actions_set.wipe_inventory") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ $t("tools.actions_set.wipe_inventory_confirm") }}
        </AlertDialogDescription>
      </AlertDialogHeader>

      <div class="space-y-2">
        <div class="flex items-center space-x-2">
          <input
            id="wipe-labels-checkbox"
            v-model="wipeLabels"
            type="checkbox"
            class="size-4 rounded border-gray-300"
          />
          <label for="wipe-labels-checkbox" class="cursor-pointer text-sm font-medium">
            {{ $t("tools.actions_set.wipe_inventory_labels") }}
          </label>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="wipe-locations-checkbox"
            v-model="wipeLocations"
            type="checkbox"
            class="size-4 rounded border-gray-300"
          />
          <label for="wipe-locations-checkbox" class="cursor-pointer text-sm font-medium">
            {{ $t("tools.actions_set.wipe_inventory_locations") }}
          </label>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="wipe-maintenance-checkbox"
            v-model="wipeMaintenance"
            type="checkbox"
            class="size-4 rounded border-gray-300"
          />
          <label for="wipe-maintenance-checkbox" class="cursor-pointer text-sm font-medium">
            {{ $t("tools.actions_set.wipe_inventory_maintenance") }}
          </label>
        </div>
      </div>

      <p class="text-sm text-gray-600">
        {{ $t("tools.actions_set.wipe_inventory_note") }}
      </p>

      <AlertDialogFooter>
        <AlertDialogCancel @click="close">
          {{ $t("global.cancel") }}
        </AlertDialogCancel>
        <Button @click="confirm">
          {{ $t("global.confirm") }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup lang="ts">
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { useDialog } from "~/components/ui/dialog-provider";
  import {
    AlertDialog,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@/components/ui/alert-dialog";
  import { Button } from "@/components/ui/button";

  const { registerOpenDialogCallback, closeDialog, addAlert, removeAlert } = useDialog();

  const dialog = ref(false);
  const wipeLabels = ref(false);
  const wipeLocations = ref(false);
  const wipeMaintenance = ref(false);

  registerOpenDialogCallback(DialogID.WipeInventory, () => {
    dialog.value = true;
    wipeLabels.value = false;
    wipeLocations.value = false;
    wipeMaintenance.value = false;
  });

  watch(
    dialog,
    val => {
      if (val) {
        addAlert("wipe-inventory-dialog");
      } else {
        removeAlert("wipe-inventory-dialog");
      }
    },
    { immediate: true }
  );

  function handleOpenChange(open: boolean) {
    if (!open) {
      close();
    }
  }

  function close() {
    dialog.value = false;
    closeDialog(DialogID.WipeInventory, undefined);
  }

  function confirm() {
    const result = {
      wipeLabels: wipeLabels.value,
      wipeLocations: wipeLocations.value,
      wipeMaintenance: wipeMaintenance.value,
    };
    dialog.value = false;
    closeDialog(DialogID.WipeInventory, result);
  }
</script>

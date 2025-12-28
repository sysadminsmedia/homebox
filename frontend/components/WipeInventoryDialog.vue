<template>
  <BaseModal v-model="dialog" max-width="600px">
    <template #title>
      <span>{{ $t("tools.actions_set.wipe_inventory") }}</span>
    </template>
    <div class="space-y-4">
      <p class="text-base">
        {{ $t("tools.actions_set.wipe_inventory_confirm") }}
      </p>
      
      <div class="space-y-2">
        <div class="flex items-center space-x-2">
          <input
            id="wipe-labels-checkbox"
            v-model="wipeLabels"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300"
          />
          <label for="wipe-labels-checkbox" class="text-sm font-medium cursor-pointer">
            {{ $t("tools.actions_set.wipe_inventory_labels") }}
          </label>
        </div>
        
        <div class="flex items-center space-x-2">
          <input
            id="wipe-locations-checkbox"
            v-model="wipeLocations"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300"
          />
          <label for="wipe-locations-checkbox" class="text-sm font-medium cursor-pointer">
            {{ $t("tools.actions_set.wipe_inventory_locations") }}
          </label>
        </div>
      </div>
      
      <p class="text-sm text-gray-600">
        {{ $t("tools.actions_set.wipe_inventory_note") }}
      </p>
    </div>
    
    <template #actions>
      <BaseButton @click="close"> {{ $t("global.cancel") }} </BaseButton>
      <BaseButton type="primary" @click="confirm">
        {{ $t("global.confirm") }}
      </BaseButton>
    </template>
  </BaseModal>
</template>

<script setup lang="ts">
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { useDialog } from "~/components/ui/dialog-provider";
  
  const { registerOpenDialogCallback, closeDialog } = useDialog();
  
  const dialog = ref(false);
  const wipeLabels = ref(false);
  const wipeLocations = ref(false);
  
  let onCloseCallback: ((result?: { wipeLabels: boolean; wipeLocations: boolean } | undefined) => void) | undefined;
  
  registerOpenDialogCallback(DialogID.WipeInventory, (params?: { onClose?: (result?: { wipeLabels: boolean; wipeLocations: boolean } | undefined) => void }) => {
    dialog.value = true;
    wipeLabels.value = false;
    wipeLocations.value = false;
    onCloseCallback = params?.onClose;
  });
  
  function close() {
    dialog.value = false;
    closeDialog(DialogID.WipeInventory, undefined);
    onCloseCallback?.(undefined);
  }
  
  function confirm() {
    dialog.value = false;
    const result = {
      wipeLabels: wipeLabels.value,
      wipeLocations: wipeLocations.value,
    };
    closeDialog(DialogID.WipeInventory, result);
    onCloseCallback?.(result);
  }
</script>

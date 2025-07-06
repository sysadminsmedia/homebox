<script setup lang="ts">
  import { DialogRoot, type DialogRootEmits, type DialogRootProps, useForwardPropsEmits } from "reka-ui";
  import { useDialog, type ActiveDialog } from "../dialog-provider/utils";

  const props = defineProps<DialogRootProps & { dialogId: string }>();
  const emits = defineEmits<DialogRootEmits>();

  const { closeDialog, activeDialog } = useDialog();

  const isOpen = computed(() => (activeDialog.value && activeDialog.value.id === props.dialogId));
  const onOpenChange = (open: boolean) => {
    if (!open) closeDialog(props.dialogId);
  };

  const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <DialogRoot v-bind="forwarded" :open="isOpen" @update:open="onOpenChange">
    <slot />
  </DialogRoot>
</template>

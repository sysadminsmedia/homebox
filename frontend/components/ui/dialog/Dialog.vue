<script setup lang="ts">
  import { DialogRoot, type DialogRootEmits, type DialogRootProps, useForwardPropsEmits } from "reka-ui";
  import { useDialog, type DialogID } from "@/components/ui/dialog-provider/utils";

  const props = defineProps<DialogRootProps & { dialogId: DialogID }>();
  const emits = defineEmits<DialogRootEmits>();

  const { closeDialog, activeDialog } = useDialog();

  const isOpen = computed(() => (activeDialog.value && activeDialog.value === props.dialogId));
  const onOpenChange = (open: boolean) => {
    if (!open) closeDialog(props.dialogId as any);
  };

  const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <DialogRoot v-bind="forwarded" :open="isOpen" @update:open="onOpenChange">
    <slot />
  </DialogRoot>
</template>

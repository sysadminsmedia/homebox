<script setup lang="ts">
  import { DialogRoot, type DialogRootEmits, type DialogRootProps, useForwardPropsEmits } from "reka-ui";
  import { useDialog, type DialogID } from "@/components/ui/dialog-provider/utils";
  import { computed, type HTMLAttributes } from "vue";

  const props = defineProps<DialogRootProps & { class?: HTMLAttributes["class"]; dialogId: DialogID }>();
  const emits = defineEmits<DialogRootEmits>();

  const { closeDialog, activeDialog } = useDialog();

  const isOpen = computed(() => (activeDialog.value && activeDialog.value === props.dialogId));
  const onOpenChange = (open: boolean) => {
    if (!open) closeDialog(props.dialogId);
  };

  const delegatedProps = computed(() => {
    const { class: _, dialogId, ...delegated } = props;

    return delegated;
  });

  const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <DialogRoot v-bind="forwarded" :open="isOpen" @update:open="onOpenChange">
    <slot />
  </DialogRoot>
</template>

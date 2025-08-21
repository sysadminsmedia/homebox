<script lang="ts" setup>
  import type { DrawerRootEmits, DrawerRootProps } from "vaul-vue";
  import { useForwardPropsEmits } from "reka-ui";
  import { DrawerRoot } from "vaul-vue";
  import { DialogID, useDialog } from "@/components/ui/dialog-provider/utils";

  const props = withDefaults(defineProps<DrawerRootProps & { dialogId: string }>(), {
    shouldScaleBackground: true,
  }) as DrawerRootProps & { dialogId: DialogID };

  const emits = defineEmits<DrawerRootEmits>();

  const { closeDialog, activeDialog } = useDialog();

  const isOpen = computed(() => activeDialog.value !== null && activeDialog.value === props.dialogId);
  const onOpenChange = (open: boolean) => {
    if (!open) closeDialog(props.dialogId as any);
  };

  const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <DrawerRoot v-bind="forwarded" :open="isOpen" @update:open="onOpenChange">
    <slot />
  </DrawerRoot>
</template>

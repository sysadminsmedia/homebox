<template>
  <AlertDialog :open="isRevealed">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ $t("global.confirm") }}</AlertDialogTitle>
        <AlertDialogDescription> {{ text }} </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="cancel(false)">
          {{ $t("global.cancel") }}
        </AlertDialogCancel>
        <AlertDialogAction @click="confirm(true)">
          {{ $t("global.confirm") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup lang="ts">
  import { useDialog } from "./ui/dialog-provider";

  const { text, isRevealed, confirm, cancel } = useConfirm();
  const { addAlert, removeAlert } = useDialog();

  watch(
    isRevealed,
    val => {
      if (val) {
        addAlert("confirm-modal");
      } else {
        removeAlert("confirm-modal");
      }
    },
    { immediate: true }
  );
</script>

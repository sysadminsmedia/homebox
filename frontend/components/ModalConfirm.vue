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
  const { text, isRevealed, confirm, cancel } = useConfirm();
</script>

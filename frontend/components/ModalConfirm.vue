<template>
  <BaseModal v-model="isRevealed" readonly @cancel="cancel(false)">
    <template #title> {{ $t("global.confirm") }} </template>
    <div>
      <p>{{ text }}</p>
    </div>
    <div class="modal-action">
      <BaseButton type="submit" @click="confirm(true)"> {{ $t("global.confirm") }} </BaseButton>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
  const { text, isRevealed, confirm, cancel } = useConfirm();
</script>

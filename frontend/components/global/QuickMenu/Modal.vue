<template>
  <BaseModal
    v-model="modal"
    :show-close-button="false"
    :click-outside-to-close="true"
    :modal-top="true"
    :class="{ 'self-start': true }"
  >
    <div class="relative">
      <span class="text-neutral-400">{{ $t("components.quick_menu.shortcut_hint") }}</span>
      <QuickMenuInput ref="inputBox" :actions="props.actions || []" @action-selected="invokeAction"></QuickMenuInput>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
  import type { ExposedProps as QuickMenuInputData, QuickMenuAction } from "./Input.vue";

  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
    actions: {
      type: Array as PropType<QuickMenuAction[]>,
      required: false,
      default: () => [],
    },
  });

  const modal = useVModel(props, "modelValue");

  const inputBox = ref<QuickMenuInputData>({ focused: false, revealActions: () => {} });

  const onModalOpen = useTimeoutFn(() => {
    inputBox.value.focused = true;
  }, 50).start;

  const onModalClose = () => {
    inputBox.value.focused = false;
  };

  watch(modal, () => (modal.value ? onModalOpen : onModalClose)());

  onStartTyping(() => {
    inputBox.value.focused = true;
  });

  function invokeAction(action: QuickMenuAction) {
    modal.value = false;
    useTimeoutFn(action.action, 100).start();
  }
</script>

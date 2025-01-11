<template>
  <BaseModal v-model="modal" :show-close-button="false">
    <div class="relative">
      <span class="text-neutral-400">{{ $t("components.quick_menu.shortcut_hint") }}</span>
      <QuickMenuInput
        ref="inputBox"
        v-model="selectedAction"
        :actions="props.actions || []"
        @quick-select="invokeAction"
      ></QuickMenuInput>
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
  const selectedAction = ref<QuickMenuAction>();

  const inputBox = ref<QuickMenuInputData>({ focused: false, revealActions: () => {} });

  const onModalOpen = useTimeoutFn(() => {
    inputBox.value.focused = true;
  }, 50).start;

  const onModalClose = () => {
    selectedAction.value = undefined;
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

  watch(selectedAction, action => {
    if (action) invokeAction(action);
  });
</script>

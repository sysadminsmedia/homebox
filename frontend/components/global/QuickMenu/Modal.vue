<template>
  <BaseModal v-model="modal" :show-close-button="false">
    <div class="relative">
      <QuickMenuInput
        ref="inputBox"
        v-model="selectedAction"
        :actions="props.actions || []"
        @quick-select="invokeAction"
      ></QuickMenuInput>
      <ul v-if="false" class="menu rounded-box w-full">
        <li v-for="(action, idx) in actions || []" :key="idx">
          <button
            class="rounded-btn w-full p-3 text-left transition-colors hover:bg-neutral hover:text-white"
            @click="invokeAction(action)"
          >
            <b v-if="action.shortcut">{{ action.shortcut }}.</b>

            {{ action.text }}
          </button>
        </li>
      </ul>
      <span class="text-base-300">Use number keys to quick select.</span>
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

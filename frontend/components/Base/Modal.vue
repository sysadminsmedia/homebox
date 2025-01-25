<template>
  <div class="z-[999]">
    <input :id="modalId" v-model="modal" type="checkbox" class="modal-toggle" />
    <div
      class="modal overflow-visible sm:modal-middle"
      :class="{ 'modal-bottom': !props.modalTop }"
      :modal-top="props.modalTop"
    >
      <div ref="modalBox" class="modal-box relative overflow-visible">
        <button
          v-if="props.showCloseButton"
          :for="modalId"
          class="btn btn-circle btn-sm absolute right-2 top-2"
          @click="close"
        >
          ✕
        </button>

        <h3 class="text-lg font-bold">
          <slot name="title"></slot>
        </h3>
        <slot> </slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  const emit = defineEmits(["cancel", "update:modelValue"]);
  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
    /**
     * in readonly mode the modal only `emits` a "cancel" event to indicate
     * that the modal was closed via the "x" button. The parent component is
     * responsible for closing the modal.
     */
    readonly: {
      type: Boolean,
      default: false,
    },
    showCloseButton: {
      type: Boolean,
      default: true,
    },
    clickOutsideToClose: {
      type: Boolean,
      default: false,
    },
    modalTop: {
      type: Boolean,
      default: false,
    },
  });

  const modalBox = ref();

  function escClose(e: KeyboardEvent) {
    if (e.key === "Escape") {
      close();
    }
  }

  if (props.clickOutsideToClose) {
    onClickOutside(modalBox, () => {
      close();
    });
  }

  function close() {
    if (props.readonly) {
      emit("cancel");
      return;
    }
    modal.value = false;
  }

  const modalId = useId();
  const modal = useVModel(props, "modelValue", emit);

  watchEffect(() => {
    if (modal.value) {
      document.addEventListener("keydown", escClose);
    } else {
      document.removeEventListener("keydown", escClose);
    }
  });
</script>

<style lang="css" scoped>
  @media (max-width: 640px) {
    .modal[modal-top] {
      align-items: start;
    }

    .modal[modal-top] :where(.modal-box) {
      max-width: none;
      --tw-translate-y: 2.5rem /* 40px */;
      --tw-scale-x: 1;
      --tw-scale-y: 1;
      transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate))
        skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
      width: 100%;
      border-top-left-radius: 0px;
      border-top-right-radius: 0px;
    }
  }
</style>

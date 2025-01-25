<template>
  <Combobox v-model="selectedAction">
    <ComboboxInput
      ref="inputBox"
      class="input input-bordered mt-2 w-full"
      @input="inputValue = $event.target.value"
    ></ComboboxInput>
    <ComboboxOptions
      class="card dropdown-content absolute max-h-48 w-full overflow-y-scroll rounded-lg border border-base-300 bg-base-100"
    >
      <ComboboxOption
        v-for="(action, idx) in filteredActions"
        :key="idx"
        v-slot="{ active }"
        :value="action"
        as="template"
      >
        <button
          class="flex w-full rounded-lg px-3 py-1.5 text-left transition-colors"
          :class="{ 'bg-primary text-primary-content': active }"
        >
          {{ action.text }}

          <kbd
            v-if="action.shortcut"
            class="kbd kbd-sm ml-auto"
            :class="{ 'border-primary-content bg-primary': active }"
          >
            {{ action.shortcut }}
          </kbd>
        </button>
      </ComboboxOption>
      <div
        v-if="filteredActions.length == 0"
        class="w-full rounded-lg p-3 text-left transition-colors hover:bg-base-300"
      >
        No actions found.
      </div>
    </ComboboxOptions>
    <ComboboxButton ref="inputBoxButton"></ComboboxButton>
  </Combobox>
</template>

<script setup lang="ts">
  import { Combobox, ComboboxInput, ComboboxOptions, ComboboxOption, ComboboxButton } from "@headlessui/vue";

  type ExposedProps = {
    focused: boolean;
    revealActions: () => void;
  };

  type QuickMenuAction = {
    text: string;
    action: () => void;
    // A character that invokes this action instantly if pressed
    shortcut?: string;
  };

  const props = defineProps({
    modelValue: {
      type: Object as PropType<QuickMenuAction>,
      required: false,
      default: undefined,
    },
    actions: {
      type: Array as PropType<QuickMenuAction[]>,
      required: true,
    },
  });

  const selectedAction = useVModel(props, "modelValue");

  const inputValue = ref("");
  const inputBox = ref();
  const inputBoxButton = ref();
  const { focused: inputBoxFocused } = useFocus(inputBox);

  const emit = defineEmits(["update:modelValue", "quickSelect"]);

  const revealActions = () => {
    unrefElement(inputBoxButton).click();
  };

  watch(inputBoxFocused, () => {
    if (inputBoxFocused.value) revealActions();
    else inputValue.value = "";
  });

  watch(inputValue, (val, oldVal) => {
    if (!oldVal) {
      const action = props.actions?.find(v => v.shortcut === val);
      if (action) {
        emit("quickSelect", action);
      }
    }
  });

  const filteredActions = computed(() => {
    const searchTerm = inputValue.value.toLowerCase();
    return (props.actions || []).filter(action => {
      return action.text.toLowerCase().includes(searchTerm) || action.shortcut?.includes(searchTerm);
    });
  });

  defineExpose({ focused: inputBoxFocused, revealActions });

  export type { QuickMenuAction, ExposedProps };
</script>

<template>
  <div v-if="!inline" class="flex w-full flex-col gap-1.5">
    <Label :for="id" class="flex w-full px-1">
      <span>{{ label }}</span>
      <span class="grow" />
      <span :class="{ 'text-destructive': isLengthInvalid }">
        {{ lengthIndicator }}
      </span>
    </Label>
    <Textarea
      :id="id"
      v-model="value"
      :placeholder="placeholder"
      class="min-h-[112px] w-full resize-none"
      @keydown="handleKeyDown"
    />
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <Label :for="id" class="flex w-full px-1 py-2">
      <span>{{ label }}</span>
      <span class="grow" />
      <span :class="{ 'text-destructive': isLengthInvalid }">
        {{ lengthIndicator }}
      </span>
    </Label>
    <Textarea
      :id="id"
      v-model="value"
      autosize
      :placeholder="placeholder"
      class="col-span-3 mt-2 w-full resize-none"
      @keydown="handleKeyDown"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed } from "vue";
  import { Label } from "~/components/ui/label";
  import { Textarea } from "~/components/ui/textarea";

  const props = defineProps({
    label: {
      type: String,
      required: true,
    },
    modelValue: {
      type: String,
      required: true,
    },
    placeholder: {
      type: String,
      default: "",
    },
    inline: {
      type: Boolean,
      default: false,
    },
    maxLength: {
      type: Number,
      default: -1,
      required: false,
    },
    minLength: {
      type: Number,
      default: -1,
      required: false,
    },
  });

  const id = useId();
  const value = useVModel(props, "modelValue");

  const isLengthInvalid = computed(() => {
    if (typeof value.value !== "string") return false;
    const len = value.value.length;
    const max = props.maxLength;
    const min = props.minLength;
    // invalid if max length exists and is exceeded OR min length exists and is not met
    return (max !== -1 && len > max) || (min !== -1 && len < min);
  });

  const lengthIndicator = computed(() => {
    if (typeof value.value !== "string") return "";
    const max = props.maxLength;
    if (max !== -1) {
      return `${value.value.length}/${max}`;
    }
    return "";
  });

  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.ctrlKey && event.key === "Enter") {
      // find the closest ancestor form element
      const targetElement = event.target as HTMLElement;
      const form = targetElement.closest("form");

      if (form) {
        event.preventDefault();
        form.requestSubmit();
      }
    }
  };
</script>

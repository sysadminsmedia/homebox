<template>
  <div v-if="!inline" class="flex w-full flex-col gap-1.5">
    <Label :for="id" class="flex w-full px-1">
      <span> {{ label }} </span>
      <span class="grow" />
      <span
        :class="{
          'text-destructive':
            typeof value === 'string' &&
            ((maxLength !== -1 && value.length > maxLength) || (minLength !== -1 && value.length < minLength)),
        }"
      >
        {{ typeof value === "string" && (maxLength !== -1 || minLength !== -1) ? `${value.length}/${maxLength}` : "" }}
      </span>
    </Label>
    <Input
      :id="id"
      ref="input"
      v-model="value"
      :placeholder="placeholder"
      :type="type"
      :required="required"
      class="w-full"
    />
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <Label class="flex w-full px-1 py-2" :for="id">
      <span> {{ label }} </span>
      <span class="grow" />
      <span
        :class="{
          'text-destructive':
            typeof value === 'string' &&
            ((maxLength !== -1 && value.length > maxLength) || (minLength !== -1 && value.length < minLength)),
        }"
      >
        {{ typeof value === "string" && (maxLength !== -1 || minLength !== -1) ? `${value.length}/${maxLength}` : "" }}
      </span>
    </Label>
    <Input
      :id="id"
      v-model="value"
      :placeholder="placeholder"
      :type="type"
      :required="required"
      class="col-span-3 mt-2 w-full"
    />
  </div>
</template>

<script lang="ts" setup>
  import { Label } from "~/components/ui/label";
  import { Input } from "~/components/ui/input";
  const props = defineProps({
    label: {
      type: String,
      default: "",
    },
    modelValue: {
      type: [String, Number],
      default: null,
    },
    required: {
      type: [Boolean],
      default: null,
    },
    type: {
      type: String,
      default: "text",
    },
    triggerFocus: {
      type: Boolean,
      default: null,
    },
    inline: {
      type: Boolean,
      default: false,
    },
    placeholder: {
      type: String,
      default: "",
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

  const input = ref<HTMLElement | null>(null);

  whenever(
    () => props.triggerFocus,
    () => {
      if (input.value) {
        input.value.focus();
      }
    }
  );

  const value = useVModel(props, "modelValue");
</script>

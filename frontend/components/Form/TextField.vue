<template>
  <div v-if="!inline" class="form-control w-full">
    <label class="label">
      <span class="label-text"> {{ label }} </span>
      <span
        :class="{
          'text-red-600':
            typeof value === 'string' &&
            ((maxLength !== -1 && value.length > maxLength) || (minLength !== -1 && value.length < minLength)),
        }"
      >
        {{ typeof value === "string" && (maxLength !== -1 || minLength !== -1) ? `${value.length}/${maxLength}` : "" }}
      </span>
    </label>
    <input
      ref="input"
      v-model="value"
      :placeholder="placeholder"
      :type="type"
      :required="required"
      class="input input-bordered w-full"
    />
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <label class="label">
      <span class="label-text"> {{ label }} </span>
      <span
        :class="{
          'text-red-600':
            typeof value === 'string' &&
            ((maxLength !== -1 && value.length > maxLength) || (minLength !== -1 && value.length < minLength)),
        }"
      >
        {{ typeof value === "string" && (maxLength !== -1 || minLength !== -1) ? `${value.length}/${maxLength}` : "" }}
      </span>
    </label>
    <input
      v-model="value"
      :placeholder="placeholder"
      :type="type"
      :required="required"
      class="input input-bordered col-span-3 mt-2 w-full"
    />
  </div>
</template>

<script lang="ts" setup>
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
      default: 0,
    },
    minLength: {
      type: Number,
      default: -1,
      required: false,
      default: Number.MAX_VALUE,
    },
  });

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

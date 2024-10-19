<template>
    <div v-if="!inline" class="form-control w-full">
      <label class="label cursor-pointer">
        <input v-model="value" type="checkbox" class="toggle toggle-primary" />
        <span class="label-text"> {{ label }}</span>
      </label>
    </div>
    <div v-else class="label cursor-pointer sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
      <label>
        <span class="label-text">
          {{ label }}
        </span>
      </label>
      <input v-model="value" type="checkbox" class="toggle toggle-primary" />
    </div>
  </template>
  
  <script setup lang="ts">
    const props = defineProps({
      modelValue: {
        type: Boolean,
        default: false,
      },
      inline: {
        type: Boolean,
        default: false,
      },
      label: {
        type: String,
        default: "",
      },
    });
  
    const value = useVModel(props, "modelValue");
  </script>
  
<template>
  <div v-if="!inline" class="flex w-full flex-col gap-1.5">
    <Label :for="id" class="flex w-full px-1">
      <span>{{ label }}</span>
      <span class="grow"></span>
      <span
        :class="{
          'text-red-600':
            typeof value === 'string' &&
            ((maxLength !== -1 && value.length > maxLength) || (minLength !== -1 && value.length < minLength)),
        }"
      >
        {{ typeof value === "string" && (maxLength !== -1 || minLength !== -1) ? `${value.length}/${maxLength}` : "" }}
      </span>
    </Label>
    <Textarea :id="id" v-model="value" :placeholder="placeholder" class="min-h-[112px] w-full resize-none" />
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <Label :for="id" class="flex w-full px-1 py-2">
      <span>{{ label }}</span>
      <span class="grow"></span>
      <span
        :class="{
          'text-red-600':
            typeof value === 'string' &&
            ((maxLength !== -1 && value.length > maxLength) || (minLength !== -1 && value.length < minLength)),
        }"
      >
        {{ typeof value === "string" && (maxLength !== -1 || minLength !== -1) ? `${value.length}/${maxLength}` : "" }}
      </span>
    </Label>
    <Textarea :id="id" v-model="value" autosize :placeholder="placeholder" class="col-span-3 mt-2 w-full resize-none" />
  </div>
</template>

<script lang="ts" setup>
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
</script>
